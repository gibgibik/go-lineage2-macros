package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"regexp"
	"strconv"
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
	item    service.ProfileTemplateItem
	lastRun time.Time
}

const (
	stackTypeMain = iota
	stackTypeSecondary
)

type stackStruct struct {
	stackType uint8
	runMutex  *sync.Mutex
	stopCh    chan struct{}
	reloadCh  chan struct{}
	waitCh    chan struct{}
	webWaitCh chan struct{}
	stack     []runStackStruct
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
	pidsStack          map[uint32]*stackStruct
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
	sendMessage(string(p))
	return 0, nil
}

func parseCondition(s string) *Condition {
	s = strings.ReplaceAll(s, "%", "")
	reg := regexp.MustCompile("(HP|MP)\\s(>|<|=)\\s(\\d+)")
	matches := reg.FindSubmatch([]byte(s))
	if len(matches) != 4 {
		return &Condition{}
	}
	value, _ := strconv.ParseFloat(string(matches[3]), 64)
	return &Condition{
		attr:  string(matches[1]),
		sign:  string(matches[2]),
		value: value,
	}
}

func initStacks(pid uint32, r *http.Request, logger *zap.SugaredLogger) error {
	profileData, err := service.GetProfileData(strings.Trim(r.RequestURI, "/"), logger)
	if err != nil {
		return err
	}
	if len(pidsStack[pid].stack) == 0 {
		for _, val := range profileData.Items {
			if val.Action == "" {
				continue
			}
			cp := pidsStack[pid]
			cp.stack = append(pidsStack[pid].stack, runStackStruct{
				item: val,
			})
			pidsStack[pid] = cp
		}

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
			cnf := cmd.Context().Value("cnf").(*core.Config)
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
		ErrorLog:     log.New(&core.FwdToZapWriter{logger}, "", 0),
		BaseContext: func(listener net.Listener) context.Context {
			return context.WithValue(ctx, "logger", logger)
		},
	}
	mux := http.NewServeMux() // Create
	mux.HandleFunc("/ws", wsHandler)
	mux.HandleFunc("/api/profile/", templateHandler)
	mux.HandleFunc("/api/start/", startHandler(ctx, cnf))
	mux.HandleFunc("/api/pause", pauseHandler())
	mux.HandleFunc("/api/stop", stopHandler())
	mux.HandleFunc("/api/init", initHandler())
	mux.HandleFunc("/api/stats", statHandler(logger))
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

func sendMessage(message string) {
	messagesStackMutex.Lock()
	messagesStack = append(messagesStack, message)
	messagesStackMutex.Unlock()
}

func randNum(min int, max int) int {
	return rand.IntN(max-min) + min
}

func createRequestError(w http.ResponseWriter, err string, code int) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(err))
}

func makeChecks(runStack map[uint32]*stackStruct, pid uint32, checksPassed bool, controlCl *service.Control, logger *zap.SugaredLogger) bool {
	if controlCl == nil {
		logger.Error("control cl is nil")
		return false
	}
	if checksPassed {
		return true
	}
	return true
}

func switchWindow(pid uint32, controlCl *service.Control, logger *zap.SugaredLogger) bool {
	curPid, err := service.GetForegroundWindowPid()
	if err != nil {
		logger.Errorf("get foreground window failed: %v", err)
		return false
	}
	if curPid == 0 || curPid == pid {
		return true
	}
	if controlCl != nil {
		controlCl.SendKey(0, "\\")
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
