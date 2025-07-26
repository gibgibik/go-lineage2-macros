package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

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

func savePresetHandler(logger *zap.SugaredLogger) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			createRequestError(writer, "Invalid Method", http.StatusMethodNotAllowed)
			return
		}
		defer request.Body.Close()
		pieces := strings.Split(request.RequestURI, "/")
		if len(pieces) != 4 {
			createRequestError(writer, "Invalid URI", http.StatusBadRequest)
			return
		}
		presetId, err := strconv.Atoi(pieces[3])
		if err != nil {
			createRequestError(writer, "Invalid URI", http.StatusBadRequest)
			return
		}
		b, err := io.ReadAll(request.Body)
		if err != nil {
			createRequestError(writer, err.Error(), http.StatusBadRequest)
			return
		}
		var pr preset.Preset
		err = json.Unmarshal(b, &pr)
		if err != nil {
			createRequestError(writer, err.Error(), http.StatusBadRequest)
			return
		}
		pr.Id = presetId
		res, err := json.Marshal(pr)
		if err != nil {
			createRequestError(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err = os.WriteFile(preset.PathPrefix+fmt.Sprintf("%d.json", presetId), res, 0600)
		if err != nil {
			createRequestError(writer, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
