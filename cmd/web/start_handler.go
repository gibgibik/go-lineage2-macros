package web

import (
	"context"
	"encoding/json"
	"image"
	"net/http"
	"time"

	"github.com/gibgibik/go-ch9329/pkg/ch9329"
	"github.com/gibgibik/go-lineage2-macros/internal/core"
	"github.com/gibgibik/go-lineage2-macros/internal/service"
	"github.com/gibgibik/go-lineage2-server/pkg/entity"
	"go.uber.org/zap"
)

func startHandler(ctx context.Context, cnf *core.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var body service.ForeGroundWindowInfo
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			createRequestError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		pid := body.Pid
		if _, ok := pidsStack[pid]; !ok {
			createRequestError(w, "Invalid PID", http.StatusBadRequest)
			return
		}
		logger := r.Context().Value("logger").(*zap.SugaredLogger).With("pid", pid)
		if !pidsStack[pid].TryLock() {
			pidsStack[pid].waitCh <- struct{}{}
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
		for k := range pidsStack {
			if k != pid {
				anotherPid = k
				break
			}
		}
		go func() {
			defer pidsStack[pid].Unlock()
			for {
				select {
				case <-ctx.Done():
					pidsStack[pid].stopCh <- struct{}{}
				case <-pidsStack[pid].reloadCh:
					cp := pidsStack[pid]
					cp.stack = nil
					pidsStack[pid] = cp
					logger.Info("reloaded")
				case <-pidsStack[pid].stopCh:
					logger.Info("macros stopped")
					cp := pidsStack[pid]
					cp.stack = nil
					pidsStack[pid] = cp
					return
				case <-pidsStack[pid].webWaitCh:
					logger.Info("pause from web context")
					<-pidsStack[pid].webWaitCh
					logger.Info("continue from web context")
				case <-pidsStack[pid].waitCh:
					logger.Info("wait start")
					controlCl.MouseActionAbsolute(ch9329.MousePressLeft, image.Point{960, 560}, 0)
					time.Sleep(time.Millisecond * 50)
					controlCl.MouseAbsoluteEnd()
					if !pidsStack[anotherPid].TryLock() {
						pidsStack[anotherPid].waitCh <- struct{}{}
					} else {
						pidsStack[anotherPid].Unlock()
					}
					<-pidsStack[pid].waitCh
					logger.Info("wait end")
					continue
				default:
					err := initStacks(body.Pid, r, logger)
					if err != nil {
						logger.Error("init stacks error: " + err.Error())
						sendMessage("init stacks error: " + err.Error())
						return
					}
					for _, profilePreset := range pidsStack[pid].stack {
						var i int
						var checksPassed bool
						var windowSwitched = false
						if pidsStack[pid].stackType == stackTypeMain {
							_ = switchWindow(pid, controlCl, logger) //switching window
						}
						for {
							if i >= len(profilePreset.item.Preset.Items) {
								break
							}
							var playerStat *entity.PlayerStat
							service.PlayerStatsMutex.Lock()
							if val, ok := service.PlayerStats.Player[pid]; ok {
								playerStat = &val
							}
							service.PlayerStatsMutex.Unlock()
							runAction := &profilePreset.item.Preset.Items[i]
							if runAction.Action == "" {
								i++
								continue
							}
							if runAction.Action == service.ActionStop {
								if runAction.LastRun.IsZero() {
									runAction.LastRun = time.Now()
								} else if runAction.PeriodMilliseconds > 0 && (runAction.LastRun.UnixMilli()+int64(runAction.PeriodMilliseconds)) < time.Now().UnixMilli() {
									if playerStat.Target.HpPercent == 0 {
										checksPassed = makeChecks(pidsStack, pid, checksPassed, controlCl, logger)
										if !checksPassed {
											logger.Error("makecheck failed")
										} else {
											if !windowSwitched && pidsStack[pid].stackType == stackTypeSecondary {
												if !pidsStack[anotherPid].TryLock() {
													pidsStack[anotherPid].waitCh <- struct{}{}
													<-pidsStack[pid].waitCh
												} else {
													pidsStack[anotherPid].Unlock()
												}
												_ = switchWindow(pid, controlCl, logger)
												windowSwitched = true
											}
											controlCl.SendKey(0, runAction.Binding)
											time.Sleep(time.Millisecond * 50)
											controlCl.EndKey()
											if runAction.DelayMilliseconds > 0 {
												time.Sleep(time.Millisecond * time.Duration(runAction.DelayMilliseconds))
											}
										}
										time.Sleep(time.Second * 10)
										pidsStack[pid].stopCh <- struct{}{}
										logger.Debug("macros stopped due to stop!!!")
									}
								}
								i++
								continue
							}
							if runAction.PeriodMilliseconds > 0 && runAction.LastRun.UnixMilli() > (time.Now().UnixMilli()-runAction.PeriodMilliseconds) {
								i++
								continue
							}
							service.PlayerStatsMutex.Lock()
							if ok, err := service.CheckCondition(runAction.ConditionsCombinator, runAction.Conditions, playerStat, service.PlayerStats.Party, logger); !ok {
								service.PlayerStatsMutex.Unlock()
								i++
								if err != nil {
									logger.Error("check condition error: " + err.Error())
								}
								continue
							} else {
								service.PlayerStatsMutex.Unlock()
							}
							if runAction.Action == service.ActionAITargetNext {
								if pidsStack[pid].stackType == stackTypeSecondary {
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
													if currentTarget != "Tarlk Basilisk" && currentTarget != "Hunter Gargoyle" && currentTarget != "Elder Tarlk Basilisk" {
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
									runAction.LastRun = time.Now()
								}
								i++
								continue
							}
							if runAction.Action == service.ActionAssistPartyMember {
								checksPassed = makeChecks(pidsStack, pid, checksPassed, controlCl, logger)
								if !checksPassed {
									logger.Error("makecheck failed")
								} else {
									if point, ok := service.AssistPartyMemberMap[runAction.Additional]; ok {
										if !windowSwitched && pidsStack[pid].stackType == stackTypeSecondary {
											if !pidsStack[anotherPid].TryLock() {
												pidsStack[anotherPid].waitCh <- struct{}{}
												<-pidsStack[pid].waitCh
											} else {
												pidsStack[anotherPid].Unlock()
											}
											windowSwitched = true
											_ = switchWindow(pid, controlCl, logger)
										}
										controlCl.MouseActionAbsolute(ch9329.MousePressRight, point, 0)
										controlCl.MouseAbsoluteEnd()
										if runAction.DelayMilliseconds > 0 {
											time.Sleep(time.Millisecond * time.Duration(runAction.DelayMilliseconds))
										}
										runAction.LastRun = time.Now()
									} else {
										logger.Error("wrong additional for assist party member: " + runAction.Additional)
									}
								}
								i++
								time.Sleep(time.Millisecond * time.Duration(randNum(50, 100)))
								continue
							}

							logger.Info("press ", runAction.Action, " ", runAction.Binding)
							if controlErr == nil {
								checksPassed = makeChecks(pidsStack, pid, checksPassed, controlCl, logger)
								if !checksPassed {
									logger.Error("makecheck failed")
								} else {
									if runAction.Action == service.ActionAttack {
										if currentTarget, _ := service.GetCurrentTarget(logger); currentTarget != "" {
											logger.Info("target is " + currentTarget)
											if currentTarget != "Tarlk Basilisk" && currentTarget != "Hunter Gargoyle" && currentTarget != "Elder Tarlk Basilisk" {
												controlCl.SendKey(0, "esc")
												time.Sleep(time.Millisecond * 50)
												controlCl.EndKey()
												time.Sleep(time.Millisecond * 50)
												i++
												continue
											}
										}
									}
									if !windowSwitched && pidsStack[pid].stackType == stackTypeSecondary {
										if !pidsStack[anotherPid].TryLock() {
											pidsStack[anotherPid].waitCh <- struct{}{}
											<-pidsStack[pid].waitCh
										} else {
											pidsStack[anotherPid].Unlock()

										}
										windowSwitched = true
										_ = switchWindow(pid, controlCl, logger)
									}
									//logger.Info("press ", runAction.item.Binding)
									controlCl.SendKey(0, runAction.Binding)
									time.Sleep(time.Millisecond * 50)
									controlCl.EndKey()
									if runAction.DelayMilliseconds > 0 {
										time.Sleep(time.Millisecond * time.Duration(runAction.DelayMilliseconds))
									}
								}
							}
							if runAction.Action == service.ActionUnstuck {
								checksPassed = makeChecks(pidsStack, pid, checksPassed, controlCl, logger)
								if !checksPassed {
									logger.Error("makecheck failed")
								} else {
									if !windowSwitched && pidsStack[pid].stackType == stackTypeSecondary {
										if !pidsStack[anotherPid].TryLock() {
											pidsStack[anotherPid].waitCh <- struct{}{}
											<-pidsStack[pid].waitCh
										} else {
											pidsStack[anotherPid].Unlock()
										}
										windowSwitched = true
										_ = switchWindow(pid, controlCl, logger)
									}
									//logger.Info("press ", runAction.item.Binding)
									controlCl.MouseActionAbsolute(ch9329.MousePressLeft, image.Point{960, 540 + 300}, 0)
									time.Sleep(time.Millisecond * 50)
									controlCl.MouseAbsoluteEnd()
									time.Sleep(time.Second * 3)
									controlCl.SendKey(0, runAction.Binding)
									time.Sleep(time.Millisecond * 50)
									controlCl.EndKey()
									time.Sleep(time.Millisecond * 50)
									controlCl.SendKey(0, "esc")
									time.Sleep(time.Millisecond * 50)
									controlCl.EndKey()
									if runAction.DelayMilliseconds > 0 {
										time.Sleep(time.Millisecond * time.Duration(runAction.DelayMilliseconds))
									}
								}
							}
							runAction.LastRun = time.Now()
							//message := fmt.Sprintf("%s %s <span style='color:red'>Target HP: [%.2f%%]</span>", runAction.item.Action, runAction.item.Binding, service.PlayerStats.Target.HpPercent)
							//logger.Info(message)
							i++
							time.Sleep(time.Millisecond * time.Duration(randNum(50, 100)))
						}
						if windowSwitched {
							windowSwitched = false
							_ = switchWindow(anotherPid, controlCl, logger)
							pidsStack[anotherPid].waitCh <- struct{}{}
						}
						//logger.Info("end interation")
						//run stack
						//time.Sleep(time.Second)
					}
				}
			}
		}()
	}
}
