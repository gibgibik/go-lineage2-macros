package cmd

import (
	"errors"
	"fmt"
	"log"
	"net/http"
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
			handle := httpServerStart(cnf, logger)
			http.HandleFunc("/ws", wsHandler(logger))
			fmt.Println("started")
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

func wsHandler(logger *zap.SugaredLogger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
}

func httpServerStart(cnf *core.Config, logger *zap.SugaredLogger) *http.Server {
	fmt.Println(cnf.WebServer.Port)
	handle := &http.Server{
		Addr:         ":" + cnf.WebServer.Port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     log.New(&core.FwdToZapWriter{logger}, "", 0),
	}
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
