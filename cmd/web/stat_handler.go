package web

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gibgibik/go-lineage2-macros/internal/service"
	"go.uber.org/zap"
)

func statHandler(logger *zap.SugaredLogger) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			createRequestError(writer, "Invalid Method", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			logger.Error("stat body read error ", err.Error())
			return
		}
		service.PlayerStatsMutex.Lock()
		defer service.PlayerStatsMutex.Unlock()
		err = json.Unmarshal(body, &service.PlayerStats)
		if err != nil {
			logger.Error("stat json unmarshal error ", err.Error())
			return
		}
	}
}
