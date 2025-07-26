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
	pathPieces := strings.SplitN(strings.Trim(r.RequestURI, "/"), "/", 4)
	if len(pathPieces) < 3 {
		logger.Infof("invalid request", strings.Trim(r.RequestURI, "/"))
		createRequestError(w, "invalid request", http.StatusBadRequest)
		return
	}
	buf, err := service.GetProfileData(pathPieces[2], logger)
	if err != nil {
		createRequestError(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, _ := json.Marshal(buf)
	w.Write(data)
}

func getProfilesListHandler(logger *zap.SugaredLogger) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			createRequestError(writer, "Invalid Method", http.StatusMethodNotAllowed)
			return
		}
		profiles := service.GetProfilesList()
		var res []*service.ProfileTemplate
		for _, profile := range profiles {
			profileData, err := service.GetProfileData(profile, logger)
			if err != nil {
				createRequestError(writer, err.Error(), http.StatusBadRequest)
				return
			}
			res = append(res, profileData)
		}
		b, err := json.Marshal(res)
		if err != nil {
			createRequestError(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Write(b)
	}
}
