package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
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
	wsConn *websocket.Conn
)

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
		if err := wsConn.WriteMessage(1, []byte(time.Now().UTC().Format(time.DateTime))); err != nil {
			logger.Errorf("Write error:", err)
			break
		}
		time.Sleep(time.Second)
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

var templateBody struct {
	Action         []string
	Details        []string
	Period_seconds []string
	Profile        string
}

func postTemplateHandler(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger) {
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
	err = json.Unmarshal(inputBody, &templateBody)
	if err != nil {
		logger.Error(err.Error())
		createRequestError(w, err.Error(), http.StatusBadRequest)
		return
	}
	reg := regexp.MustCompile("[^\\w /]")
	for idx, action := range templateBody.Action {
		if action == "" {
			continue
		}
		if _, ok := availableActions[action]; !ok {
			logger.Error(fmt.Sprintf("action %s not found, idx: %d", action, idx))
			createRequestError(w, fmt.Sprintf("action '%s' not found", action), http.StatusBadRequest)
			return
		}
		if idx > len(templateBody.Details) {
			logger.Error(fmt.Sprintf("detail %s not found, idx: %d", action, idx))
			createRequestError(w, fmt.Sprintf("detail '%s' not found", action), http.StatusBadRequest)
			return
		}
		if idx > len(templateBody.Period_seconds) {
			logger.Error(fmt.Sprintf("period seconds %s not found, idx: %d", action, idx))
			createRequestError(w, fmt.Sprintf("period seconds '%s' not found", action), http.StatusBadRequest)
			return
		}
		if (action == "/target" || action == "/delay" || action == "/useskill" || action == "/press") && len(templateBody.Details[idx]) == 0 {
			logger.Error(fmt.Sprintf("empty details: %s, idx: %d", action, idx))
			createRequestError(w, fmt.Sprintf("empty details %s", action), http.StatusBadRequest)
			return
		}
		templateBody.Action[idx] = reg.ReplaceAllString(action, "")
		templateBody.Details[idx] = reg.ReplaceAllString(templateBody.Details[idx], "")
		templateBody.Period_seconds[idx] = reg.ReplaceAllString(templateBody.Period_seconds[idx], "")
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
}
func getTemplateHandler(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger) {
	path := strings.Trim(r.RequestURI, "/")
	pathPieces := strings.SplitN(path, "/", 4)
	if len(pathPieces) < 3 {
		logger.Infof("invalid request", path)
		createRequestError(w, "invalid request", http.StatusBadRequest)
		return
	}
	fileName := getProfilePath(pathPieces[2])
	fh, err := os.OpenFile(fileName, os.O_RDWR, 0600)
	if errors.Is(err, os.ErrNotExist) {
		createRequestError(w, "file does not exist", http.StatusNotFound)
		return
	}
	buf, err := io.ReadAll(fh)
	if err != nil {
		createRequestError(w, "file read error", http.StatusInternalServerError)
	}
	w.Write(buf)
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
