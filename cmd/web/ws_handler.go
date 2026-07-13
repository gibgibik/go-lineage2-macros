package web

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value(CtxKeyLogger).(*zap.SugaredLogger)
	// Upgrade HTTP connection to WebSocket
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("Upgrade error: %v", err)
		return
	}
	defer wsConn.Close()

	// Set up pong handler to refresh read deadline on client activity
	wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))
	wsConn.SetPongHandler(func(string) error {
		wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Read from client in a separate goroutine to detect disconnection
	closeCh := make(chan struct{})
	go func() {
		defer close(closeCh)
		for {
			if _, _, err := wsConn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// Ping ticker to keep connection alive and detect dead clients
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-closeCh:
			return
		case <-pingTicker.C:
			if err := wsConn.WriteMessage(9, []byte{}); err != nil { // 9 = PingMessage
				return
			}
		default:
			if len(messagesStack) > 0 {
				messagesStackMutex.Lock()
				data, _ := json.Marshal(messagesStack)
				messagesStack = []string{}
				messagesStackMutex.Unlock()
				if err := wsConn.WriteMessage(1, data); err != nil {
					logger.Errorf("Write error: %v", err)
					return
				}
			}
			time.Sleep(time.Second)
		}
	}
}
