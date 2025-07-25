package web

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

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
			//message := []byte(fmt.Sprintf("[%s]: %s", time.Now().UTC().Format(time.DateTime), poppedElement))
			data, _ := json.Marshal(messagesStack)
			messagesStack = []string{}
			messagesStackMutex.Unlock()
			if err := wsConn.WriteMessage(1, data); err != nil {
				logger.Errorf("Write error:", err)
				break
			}
		}
		time.Sleep(time.Second)
	}
}
