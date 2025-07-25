package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gibgibik/go-lineage2-macros/internal/service"
	"go.uber.org/zap"
)

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

func postTemplateHandler(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger) {
	err := service.SaveProfileData(r.Body, logger)
	if err != nil {
		createRequestError(w, err.Error(), http.StatusBadRequest)
		return
	}
	for k := range pidsStack {
		if !pidsStack[k].TryLock() {
			pidsStack[k].reloadCh <- struct{}{}
		} else {
			pidsStack[k].Unlock()
		}
	}
}
func getTemplateHandler(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger) {
	buf, err := service.GetProfileData(strings.Trim(r.RequestURI, "/"), logger)
	if err != nil {
		createRequestError(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, _ := json.Marshal(buf)
	w.Write(data)
}
