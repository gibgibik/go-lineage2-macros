package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gibgibik/go-lineage2-macros/internal/core"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

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
	wsConn         *websocket.Conn
	runMutex       sync.Mutex
	stopRunChannel = make(chan interface{}, 1)
	stackLock      sync.Mutex
	runStack       []struct {
		action  string
		binding string
	}
	delayStack []struct {
		action       string
		binding      string
		delaySeconds int
		lastRun      time.Time
	}
	messagesStack      []string
	messagesStackMutex sync.Mutex
)

func initStacks(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger) error {
	profileData, err := getProfileData(w, r, logger)
	if err != nil {
		return err
	}
	if len(runStack) == 0 {
		for idx, val := range profileData.Actions {
			if val == "" {
				continue
			}
			if profileData.Period_seconds[idx] > 0 {
				continue
			}
			runStack = slices.Insert(runStack, 0, struct {
				action  string
				binding string
			}{
				action:  val,
				binding: profileData.Bindings[idx],
			})
		}

	}
	if len(delayStack) == 0 {
		for idx, val := range profileData.Actions {
			if val == "" {
				continue
			}
			if profileData.Period_seconds[idx] == 0 {
				continue
			}

			delayStack = slices.Insert(delayStack, 0, struct {
				action       string
				binding      string
				delaySeconds int
				lastRun      time.Time
			}{
				action:       val,
				binding:      profileData.Bindings[idx],
				delaySeconds: profileData.Period_seconds[idx],
				lastRun:      time.Time{},
			})
		}
	}
	if len(runStack) == 0 && len(delayStack) == 0 {
		logger.Error("no actions available")
		return errors.New("no actions available")
	}
	return nil
}
func createWebServerCommand(logger *zap.SugaredLogger) *cobra.Command {
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

func wsHandler(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*zap.SugaredLogger)
	// Upgrade HTTP connection to WebSocket
	//r.Header.Add("Upgrade", "websocket")
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("Upgrade error:", err)
		return
	}
	defer wsConn.Close()

	// Set custom read deadline
	wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Simple echo loop
	for {
		if len(messagesStack) > 0 {
			messagesStackMutex.Lock()
			poppedElement := messagesStack[len(messagesStack)-1] // Get the last element
			messagesStack = messagesStack[:len(messagesStack)-1] // Re-slice to exclude the last element
			messagesStackMutex.Unlock()
			if err := wsConn.WriteMessage(1, []byte(fmt.Sprintf("[%s]: %s", time.Now().UTC().Format(time.DateTime), poppedElement))); err != nil {
				logger.Errorf("Write error:", err)
				break
			}
		}
		time.Sleep(time.Millisecond * 100)
		//mt, message, err := wsConn.ReadMessage()
		//if err != nil {
		//	logger.Errorf("Read error:", err)
		//	break
		//}
		//log.Printf("Received: %s", message)
		//if err := wsConn.WriteMessage(mt, message); err != nil {
		//	logger.Errorf("Write error:", err)
		//	break
		//}
	}
}

func withCORS(next http.Handler, logger *zap.SugaredLogger) http.Handler {
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
	fmt.Println(cnf.WebServer.Port)
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
	mux.HandleFunc("/api/start/", startHandler)
	handle.Handler = withCORS(mux, logger)
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

func startHandler(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*zap.SugaredLogger)
	if !runMutex.TryLock() {
		logger.Error("already running")
		createRequestError(w, "already running", http.StatusServiceUnavailable)
		return
	}
	go func() {
		defer runMutex.Unlock()
		logger.Info("starting run")
		for {
			select {
			case <-stopRunChannel:
				logger.Info("stopping run")
				return
			default:
				stackLock.Lock()
				err := initStacks(w, r, logger)
				if err != nil {
					logger.Error("init stacks error: " + err.Error())
					sendMessage("init stacks error: " + err.Error())
					stackLock.Unlock()
					return
				}
				for idx, delayedAction := range delayStack {
					if delayedAction.lastRun.IsZero() {
						sendMessage(fmt.Sprintf("call delayed [%s] %s", delayedAction.action, delayedAction.binding))
						delayStack[idx].lastRun = time.Now().Add(time.Duration(delayedAction.delaySeconds) * time.Second)
						continue
					}
					if delayedAction.lastRun.Unix() <= time.Now().Unix() {
						sendMessage(fmt.Sprintf("call delayed [%s] %s", delayedAction.action, delayedAction.binding))
						delayStack[idx].lastRun = time.Now().Add(time.Duration(delayedAction.delaySeconds) * time.Second)
					}
				}
				for idx, runAction := range runStack {
					sendMessage(fmt.Sprintf("call [%s] %s", runAction.action, runAction.binding))
					runStack = append(runStack[:idx], runStack[idx+1:]...)
				}
				stackLock.Unlock()
				//run stack
				time.Sleep(time.Millisecond * time.Duration(randNum(500, 800)))
			}
		}
	}()
}

func sendMessage(message string) {
	messagesStackMutex.Lock()
	messagesStack = append(messagesStack, message)
	messagesStackMutex.Unlock()
}

func randNum(min int, max int) int {
	return rand.IntN(max-min) + min
}

func templateHandler(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*zap.SugaredLogger)
	if r.Method == "GET" {
		getTemplateHandler(w, r, logger)
		return
	}
	if r.Method == "POST" {
		postTemplateHandler(w, r, logger)
		return
	}
}

type templateBodyStruct struct {
	Actions        []string
	Bindings       []string
	Period_seconds []int
	Profile        string
}

func postTemplateHandler(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger) {
	stackLock.Lock()
	stackLock.Unlock()
	availableActions := map[string]interface{}{
		"/assist":   nil,
		"/attack":   nil,
		"/target":   nil,
		"/delay":    nil,
		"/useskill": nil,
		"/press":    nil,
	}
	inputBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error(err.Error())
		createRequestError(w, err.Error(), http.StatusBadRequest)
		return
	}
	var templateBody templateBodyStruct
	err = json.Unmarshal(inputBody, &templateBody)
	if err != nil {
		logger.Error(err.Error())
		createRequestError(w, err.Error(), http.StatusBadRequest)
		return
	}
	reg := regexp.MustCompile("[^\\w /]")
	for idx, action := range templateBody.Actions {
		if action == "" {
			continue
		}
		if _, ok := availableActions[action]; !ok {
			logger.Error(fmt.Sprintf("action %s not found, idx: %d", action, idx))
			createRequestError(w, fmt.Sprintf("action '%s' not found", action), http.StatusBadRequest)
			return
		}
		if idx > len(templateBody.Bindings) {
			logger.Error(fmt.Sprintf("binding %s not found, idx: %d", action, idx))
			createRequestError(w, fmt.Sprintf("binding '%s' not found", action), http.StatusBadRequest)
			return
		}
		if idx > len(templateBody.Period_seconds) {
			logger.Error(fmt.Sprintf("period seconds %s not found, idx: %d", action, idx))
			createRequestError(w, fmt.Sprintf("period seconds '%s' not found", action), http.StatusBadRequest)
			return
		}
		if (action == "/target" || action == "/delay" || action == "/useskill" || action == "/press") && len(templateBody.Bindings[idx]) == 0 {
			logger.Error(fmt.Sprintf("empty details: %s, idx: %d", action, idx))
			createRequestError(w, fmt.Sprintf("empty details %s", action), http.StatusBadRequest)
			return
		}
		templateBody.Actions[idx] = reg.ReplaceAllString(action, "")
		templateBody.Bindings[idx] = reg.ReplaceAllString(templateBody.Bindings[idx], "")
		templateBody.Period_seconds[idx] = templateBody.Period_seconds[idx]
	}
	fileName := getProfilePath(templateBody.Profile)
	if err != nil {
		logger.Error(err.Error())
		createRequestError(w, err.Error(), http.StatusInternalServerError)
	}
	tb, err := json.Marshal(templateBody)
	if err != nil {
		logger.Error(err.Error())
		createRequestError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = os.WriteFile(fileName, tb, 0600)
	if err != nil {
		logger.Error(err.Error())
		createRequestError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	runStack = []struct {
		action  string
		binding string
	}{}
	delayStack = []struct {
		action       string
		binding      string
		delaySeconds int
		lastRun      time.Time
	}{}
}
func getTemplateHandler(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger) {
	buf, err := getProfileData(w, r, logger)
	if err != nil {
		createRequestError(w, "file read error", http.StatusInternalServerError)
	}
	data, _ := json.Marshal(buf)
	w.Write(data)
}

func getProfileData(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger) (*templateBodyStruct, error) {
	path := strings.Trim(r.RequestURI, "/")
	pathPieces := strings.SplitN(path, "/", 4)
	if len(pathPieces) < 3 {
		logger.Infof("invalid request", path)
		createRequestError(w, "invalid request", http.StatusBadRequest)
		return nil, errors.New("invalid request")
	}
	fileName := getProfilePath(pathPieces[2])
	fh, err := os.OpenFile(fileName, os.O_RDWR, 0600)
	if errors.Is(err, os.ErrNotExist) {
		createRequestError(w, "file does not exist", http.StatusNotFound)
		return nil, errors.New("file does not exist")
	}
	buf, err := io.ReadAll(fh)
	var templateBody *templateBodyStruct
	err = json.Unmarshal(buf, &templateBody)
	if err != nil {
		return nil, err
	}
	return templateBody, err
}

func getProfilePath(profileName string) string {
	reg := regexp.MustCompile("\\W")
	fileName := "var/profiles/" + reg.ReplaceAllString(profileName, "") + ".json" //@todo move to config
	return fileName
}

func createRequestError(w http.ResponseWriter, err string, code int) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(err))
}
