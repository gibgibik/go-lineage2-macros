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
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type runStackStruct struct {
	item    service.ProfileTemplateItem
	lastRun time.Time
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
	runMutex           sync.Mutex
	stopRunChannel     = make(chan interface{}, 1)
	stackLock          sync.Mutex
	runStack           []runStackStruct
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

func initStacks(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger) error {
	profileData, err := service.GetProfileData(strings.Trim(r.RequestURI, "/"), logger)
	if err != nil {
		return err
	}
	if len(runStack) == 0 {
		for _, val := range profileData.Items {
			if val.Action == "" {
				continue
			}
			runStack = append(runStack, runStackStruct{
				item: val,
			})
		}

	}
	if len(runStack) == 0 {
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
			go service.StartPlayerStatUpdate(cmd.Context(), cnf.PlayerStatUrl, logger)
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
		//mt, message, err := wsConn.ReadMessage()
		//if err != nil {
		//	logger.Errorf("Read error:", err)
		//	break
		//}
		//log.Printf("Received: %s", message)
		//if err := wsConn.WriteMessage(mt, message); err != nil {
		//	logger.Errorf("Write error:", err)
		//	break
		//}
	}
}

func withCORS(next http.Handler, logger *zap.SugaredLogger) http.Handler {
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
	mux.HandleFunc("/api/stop", func(writer http.ResponseWriter, request *http.Request) {
		stopRunChannel <- struct{}{}
	})
	mux.HandleFunc("/api/init", func(writer http.ResponseWriter, request *http.Request) {
		lockResult := runMutex.TryLock()
		if lockResult {
			defer runMutex.Unlock()
		}
		response := struct {
			IsMacrosRunning bool     `json:"isMacrosRunning"`
			ProfilesList    []string `json:"profilesList"`
		}{
			IsMacrosRunning: !lockResult,
			ProfilesList:    service.GetProfilesList(),
		}
		res, _ := json.Marshal(response)
		writer.Write(res)
	})
	mux.Handle("/", http.FileServer(http.Dir("./web/dist")))
	handle.Handler = withCORS(mux, logger)
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
		logger := r.Context().Value("logger").(*zap.SugaredLogger)
		if !runMutex.TryLock() {
			logger.Error("already running")
			createRequestError(w, "already running", http.StatusServiceUnavailable)
			return
		}
		controlCl, controlErr := service.NewControl(cnf.Control)
		if controlErr != nil {
			logger.Errorf("control create failed: %v", controlErr)
		} else {
			//defer controlCl.Cl.Port.Close()
		}
		go func() {
			for {
				select {
				case <-ctx.Done():
					stopRunChannel <- struct{}{}
				case <-stopRunChannel:
					logger.Debug("macros stopped")
					stackLock.Lock()
					runStack = []runStackStruct{}
					stackLock.Unlock()
					runMutex.Unlock()
					if controlErr == nil {
						controlCl.Cl.Port.Close()
					}
					return
				default:
					stackLock.Lock()
					err := initStacks(w, r, logger)
					if err != nil {
						logger.Error("init stacks error: " + err.Error())
						sendMessage("init stacks error: " + err.Error())
						stackLock.Unlock()
						runMutex.Unlock()
						return
					}
					var i int
					for {
						if i >= len(runStack) {
							break
						}
						runAction := runStack[i]
						if runAction.item.Action == service.ActionStop {
							if runAction.lastRun.Unix() < 0 {
								runStack[i].lastRun = time.Now()
							} else if runAction.item.PeriodSeconds > 0 && (runAction.lastRun.Unix()+int64(runAction.item.PeriodSeconds)) < time.Now().Unix() {
								if service.PlayerStat.Target.HpPercent > 0 {
									time.Sleep(time.Second * 10)
								}
								controlCl.Cl.SendKey(0, runAction.item.Binding)
								controlCl.Cl.EndKey()
								time.Sleep(time.Millisecond * 100)
								stopRunChannel <- struct{}{}
								logger.Debug("macros stopped due to stop!!!")
							}
							i++
							continue
						}
						if runAction.item.PeriodSeconds > 0 && runAction.lastRun.Unix() > (time.Now().Unix()-int64(runAction.item.PeriodSeconds)) {
							i++
							continue
						}
						if ok, err := service.CheckCondition(runAction.item.ConditionsCombinator, runAction.item.Conditions, service.PlayerStat); !ok {
							i++
							if err != nil {
								logger.Error("check condition error: " + err.Error())
							}
							continue
						}
						if runAction.item.Action == service.ActionAITargetNext {
							bounds, err := service.FindBounds(cnf.BoundsUrl, logger)
							if err != nil {
								logger.Error("find bounds error: " + err.Error())
								i++
								continue
							} else {
								if controlErr == nil {
									controlCl.Cl.SendKey(ch9329.ModLeftShift, "z") //stay
									for _, bound := range bounds {
										if service.PlayerStat.Target.HpPercent > 0 {
											break
										}
										controlCl.Cl.MouseActionAbsolute(ch9329.MousePressLeft, image.Point{
											X: int((bound[2]-bound[0])/2) + bound[0],
											Y: bound[1] + 30,
										}, 0)
										controlCl.Cl.MouseAbsoluteEnd()
										//time.Sleep(time.Millisecond * time.Duration(700))
										time.Sleep(time.Millisecond * time.Duration(randNum(50, 100)))
									}
									controlCl.Cl.EndKey()
									if service.PlayerStat.Target.HpPercent == 0 {
										controlCl.Cl.MouseActionAbsolute(ch9329.MousePressRight, image.Pt(480, 320), 0)
										controlCl.Cl.MouseActionAbsolute(ch9329.MousePressRight, image.Pt(580, 320), 0)
										controlCl.Cl.MouseAbsoluteEnd()
										time.Sleep(time.Second)
									}
								}
							}
							runStack[i].lastRun = time.Now()
							i++
							continue
						}
						if runAction.item.Action == service.ActionAssistPartyMember {
							if point, ok := service.AssistPartyMemberMap[runAction.item.Additional]; ok {
								if controlErr == nil {
									controlCl.Cl.MouseActionAbsolute(ch9329.MousePressRight, point, 0)
									controlCl.Cl.MouseAbsoluteEnd()
								}
								runStack[i].lastRun = time.Now()
								//@todo need delay?
							} else {
								logger.Error("wrong additional for assist party member: " + runAction.item.Additional)
							}
							i++
							time.Sleep(time.Millisecond * time.Duration(randNum(50, 100)))
							continue
						}

						if controlErr == nil {
							controlCl.Cl.SendKey(0, runAction.item.Binding)
							controlCl.Cl.EndKey()
						}
						runStack[i].lastRun = time.Now()
						//message := fmt.Sprintf("%s %s <span style='color:red'>Target HP: [%.2f%%]</span>", runAction.item.Action, runAction.item.Binding, service.PlayerStat.Target.HpPercent)
						//logger.Info(message)
						i++
						time.Sleep(time.Millisecond * time.Duration(randNum(50, 100)))
					}
					stackLock.Unlock()
					//run stack
					time.Sleep(time.Millisecond * time.Duration(randNum(200, 300)))
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
	stackLock.Lock()
	stackLock.Unlock()
	err := service.SaveProfileData(r.Body, logger)
	if err != nil {
		createRequestError(w, err.Error(), http.StatusBadRequest)
		return
	}
	runStack = []runStackStruct{}
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
