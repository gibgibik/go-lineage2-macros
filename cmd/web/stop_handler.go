package web

import (
	"encoding/json"
	"net/http"
)

func stopHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		var pb pidBody
		defer request.Body.Close()
		if err := json.NewDecoder(request.Body).Decode(&pb); err != nil {
			createRequestError(writer, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if _, ok := pidsStack[pb.Pid]; !ok {
			createRequestError(writer, "Invalid PID", http.StatusBadRequest)
			return
		}
		if !pidsStack[pb.Pid].runMutex.TryLock() {
			pidsStack[pb.Pid].stopCh <- struct{}{}
		} else {
			pidsStack[pb.Pid].runMutex.Unlock()
		}
	}
}
