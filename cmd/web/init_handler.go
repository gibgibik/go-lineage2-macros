package web

import (
	"encoding/json"
	"net/http"

	"github.com/gibgibik/go-lineage2-macros/internal/service"
)

func initHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		initData, _ := service.Init()
		response := struct {
			RunningMacrosState map[uint32]bool `json:"runningMacrosState"`
			ProfilesList       []string        `json:"profilesList"`
			PidsData           map[uint32]string
		}{
			RunningMacrosState: make(map[uint32]bool),
			ProfilesList:       service.GetProfilesList(),
			PidsData:           initData.PidsData,
		}
		var minPid uint32
		for pid := range response.PidsData {
			if minPid == 0 || minPid > pid {
				minPid = pid
			}
		}
		if len(pidsStack) == 0 {
			pidsStack = make(map[uint32]*pidStack, 0)
			for pid := range response.PidsData {
				str := pidStack{stack: []runStackStruct{}, stopCh: make(chan struct{}), reloadCh: make(chan struct{}), waitCh: make(chan struct{}), webWaitCh: make(chan struct{})}
				if minPid == pid {
					str.stackType = stackTypeMain
				} else {
					str.stackType = stackTypeSecondary
				}
				pidsStack[pid] = &str
			}
		} else {
			for pid := range pidsStack {
				if !pidsStack[pid].TryLock() {
					response.RunningMacrosState[pid] = true
				} else {
					response.RunningMacrosState[pid] = false
					pidsStack[pid].Unlock()
				}
			}
		}
		res, _ := json.Marshal(response)
		writer.Write(res)
	}
}
