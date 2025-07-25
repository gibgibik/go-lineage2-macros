package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gibgibik/go-ch9329/pkg/ch9329"
	"github.com/gibgibik/go-lineage2-macros/internal/core"
	"github.com/gibgibik/go-lineage2-macros/internal/service"
	"github.com/gibgibik/go-lineage2-server/pkg/entity"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type runStackStruct struct {
	sync.Mutex
	item    service.ProfileTemplateItem
	lastRun time.Time
}

const (
	stackTypeMain = iota
	stackTypeSecondary
)

type stackStruct struct {
	stackType uint8
	runMutex  *sync.Mutex
	stopCh    chan struct{}
	reloadCh  chan struct{}
	waitCh    chan struct{}
	webWaitCh chan struct{}
	stack     []runStackStruct
}

var (
	startResult = make(chan error, 1)
	upgrader    = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Allow all connections by skipping origin check (NOT RECOMMENDED for production)
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	runStack           map[uint32]*stackStruct
	messagesStack      []string
	messagesStackMutex sync.Mutex
)

type WsSender interface {
	io.Writer
	Sync() error
}

type Condition struct {
	attr  string
	sign  string
	value float64
}
type BaseWsSender struct{}

type pidBody struct {
	Pid uint32 `json:"pid"`
}

func (ws BaseWsSender) Sync() error {
	return nil
}
func (ws BaseWsSender) Write(p []byte) (n int, err error) {
	sendMessage(string(p))
	return 0, nil
}

func parseCondition(s string) *Condition {
	s = strings.ReplaceAll(s, "%", "")
	reg := regexp.MustCompile("(HP|MP)\\s(>|<|=)\\s(\\d+)")
	matches := reg.FindSubmatch([]byte(s))
	if len(matches) != 4 {
		return &Condition{}
	}
	value, _ := strconv.ParseFloat(string(matches[3]), 64)
	return &Condition{
		attr:  string(matches[1]),
		sign:  string(matches[2]),
		value: value,
	}
}

func initStacks(pid uint32, r *http.Request, logger *zap.SugaredLogger) error {
	profileData, err := service.GetProfileData(strings.Trim(r.RequestURI, "/"), logger)
	if err != nil {
		return err
	}
	if len(runStack[pid].stack) == 0 {
		for _, val := range profileData.Items {
			if val.Action == "" {
				continue
			}
			cp := runStack[pid]
			cp.stack = append(runStack[pid].stack, runStackStruct{
				item: val,
			})
			runStack[pid] = cp
		}

	}
	if len(runStack[pid].stack) == 0 {
		logger.Error("no actions available")
		return errors.New("no actions available")
	}
	return nil
}
func createWebServerCommand(logger *zap.SugaredLogger) *cobra.Command {
	var webServer = &cobra.Command{
		Use: "web-server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cnf := cmd.Context().Value("cnf").(*core.Config)
			handle := httpServerStart(cmd.Context(), cnf, logger)
			for {
				select {
				case <-cmd.Context().Done():
					err := handle.Shutdown(cmd.Context())
					logger.Info(fmt.Sprintf("web-server stop result: %v", err))
					return err
				case <-startResult:
					return nil
				default:
					time.Sleep(time.Microsecond * 100000)
				}
			}
		},
	}
	return webServer
}

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

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // ✅ дозволяє всі домени (НЕБЕЗПЕЧНО для production!)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Обробка preflight-запиту
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func httpServerStart(ctx context.Context, cnf *core.Config, logger *zap.SugaredLogger) *http.Server {
	logger.Debug("starting webserver on port :", cnf.WebServer.Port)
	handle := &http.Server{
		Addr:         ":" + cnf.WebServer.Port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     log.New(&core.FwdToZapWriter{logger}, "", 0),
		BaseContext: func(listener net.Listener) context.Context {
			return context.WithValue(ctx, "logger", logger)
		},
	}
	mux := http.NewServeMux() // Create
	mux.HandleFunc("/ws", wsHandler)
	mux.HandleFunc("/api/profile/", templateHandler)
	mux.HandleFunc("/api/start/", startHandler(ctx, cnf))
	mux.HandleFunc("/api/pause", func(writer http.ResponseWriter, request *http.Request) {
		var pb pidBody
		defer request.Body.Close()
		if err := json.NewDecoder(request.Body).Decode(&pb); err != nil {
			createRequestError(writer, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if _, ok := runStack[pb.Pid]; !ok {
			createRequestError(writer, "Invalid PID", http.StatusBadRequest)
			return
		}
		if !runStack[pb.Pid].runMutex.TryLock() {
			runStack[pb.Pid].webWaitCh <- struct{}{}
		} else {
			runStack[pb.Pid].runMutex.Unlock()
		}
	})
	mux.HandleFunc("/api/stop", func(writer http.ResponseWriter, request *http.Request) {
		var pb pidBody
		defer request.Body.Close()
		if err := json.NewDecoder(request.Body).Decode(&pb); err != nil {
			createRequestError(writer, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if _, ok := runStack[pb.Pid]; !ok {
			createRequestError(writer, "Invalid PID", http.StatusBadRequest)
			return
		}
		if !runStack[pb.Pid].runMutex.TryLock() {
			runStack[pb.Pid].stopCh <- struct{}{}
		} else {
			runStack[pb.Pid].runMutex.Unlock()
		}
	})
	mux.HandleFunc("/api/init", func(writer http.ResponseWriter, request *http.Request) {
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
		if len(runStack) == 0 {
			runStack = make(map[uint32]*stackStruct, 0)
			for pid := range response.PidsData {
				str := stackStruct{runMutex: &sync.Mutex{}, stack: []runStackStruct{}, stopCh: make(chan struct{}), reloadCh: make(chan struct{}), waitCh: make(chan struct{}), webWaitCh: make(chan struct{})}
				if minPid == pid {
					str.stackType = stackTypeMain
				} else {
					str.stackType = stackTypeSecondary
				}
				runStack[pid] = &str
			}
		} else {
			for pid := range runStack {
				if !runStack[pid].runMutex.TryLock() {
					response.RunningMacrosState[pid] = true
				} else {
					response.RunningMacrosState[pid] = false
					runStack[pid].runMutex.Unlock()
				}
			}
		}
		res, _ := json.Marshal(response)
		writer.Write(res)
	})
	mux.HandleFunc("/api/stats", func(writer http.ResponseWriter, request *http.Request) {
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
	})
	mux.Handle("/", http.FileServer(http.Dir("./web/dist")))
	handle.Handler = withCORS(mux)
	go func() {
		if err := handle.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error("http server fatal error: " + err.Error())
			}
			startResult <- err
		}
	}()

	return handle
}

func startHandler(ctx context.Context, cnf *core.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var body service.ForeGroundWindowInfo
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			createRequestError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		pid := body.Pid
		if _, ok := runStack[pid]; !ok {
			createRequestError(w, "Invalid PID", http.StatusBadRequest)
			return
		}
		logger := r.Context().Value("logger").(*zap.SugaredLogger).With("pid", pid)
		if !runStack[pid].runMutex.TryLock() {
			runStack[pid].waitCh <- struct{}{}
			logger.Error("already running")
			createRequestError(w, "already running", http.StatusServiceUnavailable)
			return
		}
		logger.Info("starting macros")
		controlCl, controlErr := service.GetControl(cnf.Control)
		if controlErr != nil {
			logger.Errorf("control create failed: %v", controlErr)
		} else {
			//defer controlCl.cl.Port.Close()
		}
		var anotherPid uint32
		for k := range runStack {
			if k != pid {
				anotherPid = k
				break
			}
		}
		go func() {
			defer runStack[pid].runMutex.Unlock()
			for {
				select {
				case <-ctx.Done():
					runStack[pid].stopCh <- struct{}{}
				case <-runStack[pid].reloadCh:
					cp := runStack[pid]
					cp.stack = []runStackStruct{}
					runStack[pid] = cp
					logger.Info("reloaded")
				case <-runStack[pid].stopCh:
					logger.Info("macros stopped")
					cp := runStack[pid]
					cp.stack = []runStackStruct{}
					runStack[pid] = cp
					return
				case <-runStack[pid].webWaitCh:
					logger.Info("pause from web context")
					<-runStack[pid].webWaitCh
					logger.Info("continue from web context")
				case <-runStack[pid].waitCh:
					logger.Info("wait start")
					controlCl.MouseActionAbsolute(ch9329.MousePressLeft, image.Point{960, 560}, 0)
					time.Sleep(time.Millisecond * 50)
					controlCl.MouseAbsoluteEnd()
					if !runStack[anotherPid].runMutex.TryLock() {
						runStack[anotherPid].waitCh <- struct{}{}
					} else {
						runStack[anotherPid].runMutex.Unlock()
					}
					<-runStack[pid].waitCh
					logger.Info("wait end")
					continue
				default:
					err := initStacks(body.Pid, r, logger)
					if err != nil {
						logger.Error("init stacks error: " + err.Error())
						sendMessage("init stacks error: " + err.Error())
						return
					}
					var i int
					var checksPassed bool
					var windowSwitched = false
					if runStack[pid].stackType == stackTypeMain {
						_ = switchWindow(pid, controlCl, logger) //switching window
					}
					for {
						if i >= len(runStack[pid].stack) {
							break
						}
						var playerStat *entity.PlayerStat
						service.PlayerStatsMutex.Lock()
						if val, ok := service.PlayerStats.Player[pid]; ok {
							playerStat = &val
						}
						service.PlayerStatsMutex.Unlock()
						runAction := &runStack[pid].stack[i]
						if runAction.item.Action == service.ActionStop {
							if runAction.lastRun.IsZero() {
								runStack[pid].stack[i].lastRun = time.Now()
							} else if runAction.item.PeriodMilliseconds > 0 && (runAction.lastRun.UnixMilli()+int64(runAction.item.PeriodMilliseconds)) < time.Now().UnixMilli() {
								if playerStat.Target.HpPercent == 0 {
									checksPassed = makeChecks(runStack, pid, checksPassed, controlCl, logger)
									if !checksPassed {
										logger.Error("makecheck failed")
									} else {
										if !windowSwitched && runStack[pid].stackType == stackTypeSecondary {
											if !runStack[anotherPid].runMutex.TryLock() {
												runStack[anotherPid].waitCh <- struct{}{}
												<-runStack[pid].waitCh
											} else {
												runStack[anotherPid].runMutex.Unlock()
											}
											_ = switchWindow(pid, controlCl, logger)
											windowSwitched = true
										}
										//logger.Info("press ", runAction.item.Binding)
										controlCl.SendKey(0, runAction.item.Binding)
										time.Sleep(time.Millisecond * 50)
										controlCl.EndKey()
										if runAction.item.DelayMilliseconds > 0 {
											time.Sleep(time.Millisecond * time.Duration(runAction.item.DelayMilliseconds))
										}
									}
									time.Sleep(time.Second * 10)
									runStack[pid].stopCh <- struct{}{}
									logger.Debug("macros stopped due to stop!!!")
								}
							}
							i++
							continue
						}
						if runAction.item.PeriodMilliseconds > 0 && runAction.lastRun.UnixMilli() > (time.Now().UnixMilli()-runAction.item.PeriodMilliseconds) {
							i++
							continue
						}
						service.PlayerStatsMutex.Lock()
						if ok, err := service.CheckCondition(runAction.item.ConditionsCombinator, runAction.item.Conditions, playerStat, service.PlayerStats.Party, logger); !ok {
							service.PlayerStatsMutex.Unlock()
							i++
							if err != nil {
								logger.Error("check condition error: " + err.Error())
							}
							continue
						} else {
							service.PlayerStatsMutex.Unlock()
						}
						if runAction.item.Action == service.ActionAITargetNext {
							if runStack[pid].stackType == stackTypeSecondary {
								logger.Error("ainexttarget isn't supported by the bot yet")
							} else {
								bounds, err := service.FindBounds(logger)
								if err != nil {
									logger.Error("find bounds error: " + err.Error())
									i++
									continue
								} else {
									if controlErr == nil {
										controlCl.SendKey(ch9329.ModLeftShift, "z") //stay
										time.Sleep(time.Millisecond * 50)
										for _, bound := range bounds.Boxes {
											if playerStat.Target.HpPercent > 0 {
												break
											}
											controlCl.MouseActionAbsolute(ch9329.MousePressLeft, image.Point{
												X: int((bound[2]-bound[0])/2) + bound[0],
												Y: bound[1] + 30,
											}, 0)
											controlCl.MouseAbsoluteEnd()
											time.Sleep(time.Millisecond * 50)
											if currentTarget, _ := service.GetCurrentTarget(logger); currentTarget != "" {
												logger.Info("target is " + currentTarget)
												if currentTarget == "Gibik" || (currentTarget != "Cave Servant" && currentTarget != "Shackle") {
													//controlCl.SendKey(0, "esc")
													time.Sleep(time.Millisecond * 50)
												} else {
													break
												}
											}
											//time.Sleep(time.Millisecond * time.Duration(randNum(400, 500)))
										}
										if playerStat.Target.HpPercent == 0 {
											controlCl.MouseActionAbsolute(ch9329.MousePressRight, image.Pt(480, 320), 0)
											controlCl.MouseActionAbsolute(ch9329.MousePressRight, image.Pt(580, 320), 0)
											controlCl.MouseAbsoluteEnd()
										}
										controlCl.EndKey()
									}
								}
								runStack[pid].stack[i].lastRun = time.Now()
							}
							i++
							continue
						}
						if runAction.item.Action == service.ActionAssistPartyMember {
							checksPassed = makeChecks(runStack, pid, checksPassed, controlCl, logger)
							if !checksPassed {
								logger.Error("makecheck failed")
							} else {
								if point, ok := service.AssistPartyMemberMap[runAction.item.Additional]; ok {
									if !windowSwitched && runStack[pid].stackType == stackTypeSecondary {
										if !runStack[anotherPid].runMutex.TryLock() {
											runStack[anotherPid].waitCh <- struct{}{}
											<-runStack[pid].waitCh
										} else {
											runStack[anotherPid].runMutex.Unlock()
										}
										windowSwitched = true
										_ = switchWindow(pid, controlCl, logger)
									}
									//logger.Info("press ", runAction.item.Binding)
									controlCl.MouseActionAbsolute(ch9329.MousePressRight, point, 0)
									controlCl.MouseAbsoluteEnd()
									if runAction.item.DelayMilliseconds > 0 {
										time.Sleep(time.Millisecond * time.Duration(runAction.item.DelayMilliseconds))
									}
									runStack[pid].stack[i].lastRun = time.Now()
									//@todo need delay?
								} else {
									logger.Error("wrong additional for assist party member: " + runAction.item.Additional)
								}
							}
							i++
							time.Sleep(time.Millisecond * time.Duration(randNum(50, 100)))
							continue
						}

						if controlErr == nil {
							checksPassed = makeChecks(runStack, pid, checksPassed, controlCl, logger)
							if !checksPassed {
								logger.Error("makecheck failed")
							} else {
								if runAction.item.Action == service.ActionAttack {
									if currentTarget, _ := service.GetCurrentTarget(logger); currentTarget != "" {
										logger.Info("target is " + currentTarget)
										if currentTarget == "Gibik" || (currentTarget != "Cave Servant" && currentTarget != "Shackle") {
											controlCl.SendKey(0, "esc")
											time.Sleep(time.Millisecond * 50)
											controlCl.EndKey()
											time.Sleep(time.Millisecond * 50)
											i++
											continue
										}
									}
								}
								if !windowSwitched && runStack[pid].stackType == stackTypeSecondary {
									if !runStack[anotherPid].runMutex.TryLock() {
										runStack[anotherPid].waitCh <- struct{}{}
										<-runStack[pid].waitCh
									} else {
										runStack[anotherPid].runMutex.Unlock()

									}
									windowSwitched = true
									_ = switchWindow(pid, controlCl, logger)
								}
								//logger.Info("press ", runAction.item.Binding)
								controlCl.SendKey(0, runAction.item.Binding)
								time.Sleep(time.Millisecond * 50)
								controlCl.EndKey()
								if runAction.item.DelayMilliseconds > 0 {
									time.Sleep(time.Millisecond * time.Duration(runAction.item.DelayMilliseconds))
								}
							}
						}
						if runAction.item.Action == service.ActionUnstuck {
							checksPassed = makeChecks(runStack, pid, checksPassed, controlCl, logger)
							if !checksPassed {
								logger.Error("makecheck failed")
							} else {
								if !windowSwitched && runStack[pid].stackType == stackTypeSecondary {
									if !runStack[anotherPid].runMutex.TryLock() {
										runStack[anotherPid].waitCh <- struct{}{}
										<-runStack[pid].waitCh
									} else {
										runStack[anotherPid].runMutex.Unlock()
									}
									windowSwitched = true
									_ = switchWindow(pid, controlCl, logger)
								}
								//logger.Info("press ", runAction.item.Binding)
								controlCl.MouseActionAbsolute(ch9329.MousePressLeft, image.Point{960, 540 + 300}, 0)
								time.Sleep(time.Millisecond * 50)
								controlCl.MouseAbsoluteEnd()
								time.Sleep(time.Second * 3)
								controlCl.SendKey(0, runAction.item.Binding)
								time.Sleep(time.Millisecond * 50)
								controlCl.EndKey()
								time.Sleep(time.Millisecond * 50)
								controlCl.SendKey(0, "esc")
								time.Sleep(time.Millisecond * 50)
								controlCl.EndKey()
								if runAction.item.DelayMilliseconds > 0 {
									time.Sleep(time.Millisecond * time.Duration(runAction.item.DelayMilliseconds))
								}
							}
						}
						runStack[pid].stack[i].lastRun = time.Now()
						//message := fmt.Sprintf("%s %s <span style='color:red'>Target HP: [%.2f%%]</span>", runAction.item.Action, runAction.item.Binding, service.PlayerStats.Target.HpPercent)
						//logger.Info(message)
						i++
						time.Sleep(time.Millisecond * time.Duration(randNum(50, 100)))
					}
					if windowSwitched {
						windowSwitched = false
						_ = switchWindow(anotherPid, controlCl, logger)
						runStack[anotherPid].waitCh <- struct{}{}
					}
					//logger.Info("end interation")
					//run stack
					time.Sleep(time.Millisecond * time.Duration(randNum(200, 300)))
					//time.Sleep(time.Second)
				}
			}
		}()
	}
}

func sendMessage(message string) {
	messagesStackMutex.Lock()
	messagesStack = append(messagesStack, message)
	messagesStackMutex.Unlock()
}

func randNum(min int, max int) int {
	return rand.IntN(max-min) + min
}

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
	for k := range runStack {
		if !runStack[k].runMutex.TryLock() {
			runStack[k].reloadCh <- struct{}{}
		} else {
			runStack[k].runMutex.Unlock()
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

func createRequestError(w http.ResponseWriter, err string, code int) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(err))
}

func makeChecks(runStack map[uint32]*stackStruct, pid uint32, checksPassed bool, controlCl *service.Control, logger *zap.SugaredLogger) bool {
	if controlCl == nil {
		logger.Error("control cl is nil")
		return false
	}
	if checksPassed {
		return true
	}
	return true
	//if runStack[pid].stackType == stackTypeSecondary {
	//	for k := range runStack {
	//		if k != pid {
	//			runStack[k].waitCh <- struct{}{}
	//			break
	//		}
	//	}
	//	<-runStack[pid].waitCh //get control
	//	return switchWindow(pid, controlCl, logger)
	//} else {
	//	return true
	//}
}

func switchWindow(pid uint32, controlCl *service.Control, logger *zap.SugaredLogger) bool {
	curPid, err := service.GetForegroundWindowPid()
	if err != nil {
		logger.Errorf("get foreground window failed: %v", err)
		return false
	}
	if curPid == 0 || curPid == pid {
		return true
	}
	if controlCl != nil {
		controlCl.SendKey(0, "\\")
		time.Sleep(time.Millisecond * 50)
		controlCl.EndKey()
		time.Sleep(time.Millisecond * 200)
	}
	curPid, err = service.GetForegroundWindowPid()
	if err != nil {
		logger.Errorf("get foreground window failed: %v", err)
		return false
	}
	if curPid != pid {
		logger.Errorf("alt tab failed, current pid is %d, window is %d", curPid, pid)
		return false
	}

	return true
}
