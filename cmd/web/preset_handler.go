package web

import (
	"encoding/json"
	"net/http"

	"github.com/gibgibik/go-lineage2-macros/internal/preset"
	"go.uber.org/zap"
)

func getPresetsListHandler(logger *zap.SugaredLogger) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			createRequestError(writer, "Invalid Method", http.StatusMethodNotAllowed)
			return
		}
		result, err := preset.GetList(logger)
		if err != nil {
			createRequestError(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		b, err := json.Marshal(result)
		if err != nil {
			createRequestError(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Write(b)
	}
}
