package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gibgibik/go-lineage2-macros/internal/core"
	"github.com/gibgibik/go-lineage2-macros/internal/service"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type runStackStruct struct {
	sync.Mutex
	item service.ProfilePreset
	//lastRun time.Time
}

const (
	stackTypeMain = iota
	stackTypeSecondary
)

const (
	messageTypeError = iota
	messageTypeText
	messageTypePlayerStats
	messageTypeTargetStats
)

type pidStack struct {
	sync.Mutex
	stackType        uint8
	preferredTargets []string
	allowedTargets   []string
	stopCh           chan struct{}
	reloadCh         chan struct{}
	waitCh           chan struct{}
	webWaitCh        chan struct{}
	stack            []*runStackStruct
}

var (
	startResult = make(chan error, 1)
	upgrader    = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Allow all connections by skipping origin check (NOT RECOMMENDED for production)
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	pidsStack          map[uint32]*pidStack
	pidsStackMu        sync.RWMutex
	messagesStack      []string
	messagesStackMutex sync.Mutex
)

type WsSender interface {
	io.Writer
	Sync() error
}

type Condition struct {
	attr  string
	sign  string
	value float64
}
type BaseWsSender struct{}

type pidBody struct {
	Pid uint32 `json:"pid"`
}

func (ws BaseWsSender) Sync() error {
	return nil
}
func (ws BaseWsSender) Write(p []byte) (n int, err error) {
	sendMessage(messageTypeText, string(p))
	return 0, nil
}

func initStacks(pid uint32, r *http.Request, logger *zap.SugaredLogger) error {
	pathPieces := strings.SplitN(strings.Trim(r.RequestURI, "/"), "/", 4)
	if len(pathPieces) < 3 {
		logger.Infof("invalid request: %s", strings.Trim(r.RequestURI, "/"))
		return errors.New("invalid request")
	}
	profileData, err := service.GetProfileData(pathPieces[2], logger)
	if err != nil {
		return err
	}
	var profilePresets []*runStackStruct
	if len(pidsStack[pid].stack) == 0 {
		for _, val := range profileData.Items {
			if !val.IsActive {
				continue
			}
			profilePresets = append(profilePresets,
				&runStackStruct{
					item: val,
				})
		}
		cp := pidsStack[pid]
		cp.stack = profilePresets
		for i, t := range profileData.AllowedTargets {
			profileData.AllowedTargets[i] = strings.ReplaceAll(strings.ToLower(t), " ", "")
		}
		cp.allowedTargets = profileData.AllowedTargets
		for i, t := range profileData.PreferredTargets {
			profileData.PreferredTargets[i] = strings.ReplaceAll(strings.ToLower(t), " ", "")
		}
		cp.preferredTargets = profileData.PreferredTargets
		pidsStack[pid] = cp
	}

	if len(pidsStack[pid].stack) == 0 {
		logger.Error("no actions available")
		return errors.New("no actions available")
	}
	return nil
}
func CreateWebServerCommand(logger *zap.SugaredLogger) *cobra.Command {
	var webServer = &cobra.Command{
		Use: "web-server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cnf := cmd.Context().Value(CtxKeyConfig).(*core.Config)
			handle := httpServerStart(cmd.Context(), cnf, logger)
			for {
				select {
				case <-cmd.Context().Done():
					err := handle.Shutdown(cmd.Context())
					logger.Info(fmt.Sprintf("web-server stop result: %v", err))
					return err
				case <-startResult:
					return nil
				default:
					time.Sleep(time.Microsecond * 100000)
				}
			}
		},
	}

	return webServer
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // ✅ дозволяє всі домени (НЕБЕЗПЕЧНО для production!)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Обробка preflight-запиту
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func httpServerStart(ctx context.Context, cnf *core.Config, logger *zap.SugaredLogger) *http.Server {
	logger.Debug("starting webserver on port :", cnf.WebServer.Port)
	handle := &http.Server{
		Addr:         ":" + cnf.WebServer.Port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     log.New(&core.FwdToZapWriter{Logger: logger}, "", 0),
		BaseContext: func(listener net.Listener) context.Context {
			return context.WithValue(ctx, CtxKeyLogger, logger)
		},
	}
	mux := http.NewServeMux() // Create
	mux.HandleFunc("/ws", wsHandler)
	mux.HandleFunc("/api/profile", getProfilesListHandler(logger))
	mux.HandleFunc("/api/profile/", templateHandler)
	mux.HandleFunc("/api/start/", startHandler(ctx, cnf))
	mux.HandleFunc("/api/pause", pauseHandler())
	mux.HandleFunc("/api/stop", stopHandler())
	mux.HandleFunc("/api/init", initHandler())
	mux.HandleFunc("/api/preset", getPresetsListHandler(logger))
	mux.HandleFunc("/api/preset/", savePresetHandler(logger))
	mux.HandleFunc("/api/stats", statHandler(logger))
	mux.HandleFunc("/api/npc", npcHandler(logger))
	mux.Handle("/", http.FileServer(http.Dir("./web/dist")))
	handle.Handler = withCORS(mux)
	go func() {
		if err := handle.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error("http server fatal error: " + err.Error())
			}
			startResult <- err
		}
	}()

	return handle
}

// trySignal performs a non-blocking send on a signalling channel.
// The pidStack channels are buffered with size 1 and are both sent to and
// received from by the worker goroutine itself as well as by HTTP handlers.
// A plain blocking send can deadlock the worker when the single buffer slot
// is already full (e.g. a stop was already queued), which froze the whole
// macro loop with no further log output. Dropping the duplicate signal is
// safe here: the pending one already conveys the same intent.
func trySignal(ch chan struct{}) bool {
	select {
	case ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// waitSignalTimeout blocks until a signal arrives on ch or the timeout
// elapses. It is used for the two-worker coordination handshake (poke the
// other worker's waitCh, then wait to be woken). Bounding the wait means a
// missed or dropped handshake signal degrades into a brief delay instead of
// freezing the worker forever.
func waitSignalTimeout(ch chan struct{}, timeout time.Duration) bool {
	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		return false
	}
}

// handshakeTimeout bounds how long a worker will wait for the partner worker
// to acknowledge a window-switch handshake.
const handshakeTimeout = 10 * time.Second

// innerControl is the decision returned by pollControlSignals telling the
// inner preset loop how to react to a control signal that arrived while it
// was busy.
type innerControl int

const (
	innerContinue innerControl = iota // no signal pending, keep running
	innerStop                         // stop/ctx-cancel: abort the whole run
	innerReload                       // profile changed: restart the stack
)

// pollControlSignals performs non-blocking checks of the per-pid control
// channels from *inside* the inner preset loop.
//
// Previously these channels were only drained at the top of the worker's
// outer select loop. While the macro was actively running it spent nearly
// all of its time inside the inner item loop, so pause/stop/reload/ctx-cancel
// signals sat in their buffered channels unobserved until the current preset
// stack finished. If the inner loop was stuck (foreground mismatch, no target,
// conditions never passing) that could be effectively forever, which is why
// "stop/pause didn't work and there were no press logs".
//
// Draining them here makes those controls take effect within one inner-loop
// iteration regardless of what the loop is doing. For pause (webWaitCh) it
// blocks until the matching resume signal arrives, mirroring the outer-loop
// behaviour.
func pollControlSignals(ctx context.Context, pid uint32, logger *zap.SugaredLogger) innerControl {
	ps := pidsStack[pid]
	select {
	case <-ctx.Done():
		return innerStop
	case <-ps.stopCh:
		logger.Info("macros stopped")
		return innerStop
	case <-ps.reloadCh:
		logger.Info("reload requested")
		return innerReload
	case <-ps.webWaitCh:
		logger.Info("pause from web context")
		// Block until resumed (or torn down) so we do not keep pressing keys
		// while paused.
		select {
		case <-ps.webWaitCh:
			logger.Info("continue from web context")
			return innerContinue
		case <-ctx.Done():
			return innerStop
		case <-ps.stopCh:
			logger.Info("macros stopped")
			return innerStop
		}
	default:
		return innerContinue
	}
}

func sendMessage(messageType int, message any) {
	messagesStackMutex.Lock()
	b, _ := json.Marshal(map[int]any{
		messageType: message,
	})
	messagesStack = append(messagesStack, string(b))
	if len(messagesStack) > 5 {
		messagesStack = messagesStack[4:]
	}
	messagesStackMutex.Unlock()
}

func randNum(min int, max int) int {
	return rand.IntN(max-min) + min
}

func createRequestError(w http.ResponseWriter, err string, code int) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(err))
}

func makeChecks(runStack map[uint32]*pidStack, pid uint32, checksPassed bool, controlCl *service.Control, logger *zap.SugaredLogger) bool {
	if controlCl == nil {
		logger.Error("control cl is nil")
		return false
	}
	if checksPassed {
		return true
	}
	curPid, err := service.GetForegroundWindowPid()
	if err != nil {
		logger.Errorf("makeChecks: get foreground window failed: %v", err)
		return false
	}
	if curPid != pid {
		logger.Errorf("makeChecks: foreground window mismatch, expected %d got %d", pid, curPid)
		return false
	}
	return true
}

func switchWindow(pid uint32, controlCl *service.Control, logger *zap.SugaredLogger) bool {
	return true
	curPid, err := service.GetForegroundWindowPid()
	if err != nil {
		logger.Errorf("get foreground window failed: %v", err)
		return false
	}
	if curPid == 0 || curPid == pid {
		return true
	}
	if controlCl != nil {
		controlCl.SendKey(0, "home")
		time.Sleep(time.Millisecond * 50)
		controlCl.EndKey()
		time.Sleep(time.Millisecond * 200)

	}
	curPid, err = service.GetForegroundWindowPid()
	if err != nil {
		logger.Errorf("get foreground window failed: %v", err)
		return false
	}
	if curPid != pid {
		logger.Errorf("alt tab failed, current pid is %d, window is %d", curPid, pid)
		return false
	}

	return true
}
