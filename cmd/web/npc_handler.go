package web

import (
	"encoding/json"
	"net/http"

	"github.com/gibgibik/go-lineage2-macros/internal/npc"
	"go.uber.org/zap"
)

func npcHandler(logger *zap.SugaredLogger) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			createRequestError(writer, "Invalid Method", http.StatusMethodNotAllowed)
			return
		}
		r, _ := json.Marshal(npc.NpcList)
		writer.Write(r)
	}
}
