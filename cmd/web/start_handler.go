package web

import (
	"context"
	"encoding/json"
	"image"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"
	"github.com/gibgibik/go-ch9329/pkg/ch9329"
	"github.com/gibgibik/go-lineage2-macros/internal/core"
	"github.com/gibgibik/go-lineage2-macros/internal/preset"
	"github.com/gibgibik/go-lineage2-macros/internal/service"
	"github.com/gibgibik/go-lineage2-server/pkg/entity"
	"go.uber.org/zap"
)

const (
	TargetNameThreshold = 1
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
		logger := r.Context().Value(CtxKeyLogger).(*zap.SugaredLogger).With("pid", pid)
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
			pidsStack[pid].Unlock()
			createRequestError(w, "control device unavailable: "+controlErr.Error(), http.StatusServiceUnavailable)
			return
		}
		initPartyMemberMap(cnf)
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
						sendMessage(messageTypeError, "init stacks error: "+err.Error())
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
								if playerStat.CP.Percent < 90 {
									service.PlayerStatsMutex.Unlock()
									pidsStack[pid].stopCh <- struct{}{}
									logger.Debug("macros stopped due to not full cp!!!")
									botToken := "6181694377:AAGI5iIFTePIoR-oqGNYAi38ygEbKK6Zqfs"
									cmd := exec.Command(
										"curl",
										"-X", "POST",
										"-H", "Content-Type: application/json",
										"-d", `{"chat_id": "451312638", "text": "STOPPED", "disable_notification": true}`,
										"https://api.telegram.org/bot"+botToken+"/sendMessage",
									)
									_ = cmd.Run()
									break
								}
							}
							service.PlayerStatsMutex.Unlock()
							runAction := &profilePreset.item.Preset.Items[i]
							if runAction.Action == "" {
								i++
								time.Sleep(time.Millisecond * 1)
								continue
							}
							if runAction.PeriodMilliseconds > 0 && runAction.LastRun.UnixMilli() > (time.Now().UnixMilli()-runAction.PeriodMilliseconds) {
								if profilePreset.item.BatchRun == true {
									break
								}
								i++
								time.Sleep(time.Millisecond * 1)
								continue
							}
							service.PlayerStatsMutex.Lock()
							if ok, err := service.CheckCondition(runAction.ConditionsCombinator, runAction.Conditions, playerStat, service.PlayerStats.Party); !ok {
								service.PlayerStatsMutex.Unlock()
								if profilePreset.item.BatchRun == true {
									break
								}
								i++
								if err != nil {
									logger.Error("check condition error: " + err.Error())
								}
								time.Sleep(time.Millisecond * 1)
								continue
							} else {
								service.PlayerStatsMutex.Unlock()
							}
							if runAction.Action == service.ActionStop {
								if !pidsStack[pid].TryLock() {
									pidsStack[pid].stopCh <- struct{}{}
								}
								return
							}
							if runAction.Action == service.ActionAITargetNext {
								if pidsStack[pid].stackType == stackTypeSecondary {
									logger.Error("ainexttarget isn't supported by the bot yet")
									i++
									continue
								} else {
									bounds, err := service.FindBounds(logger)
									logger.Info("find bounds ", bounds)
									if err != nil {
										logger.Error("find bounds error: " + err.Error())
										i++
										time.Sleep(time.Millisecond * 1)
										continue
									} else {
										_, _ = controlCl.SendKey(ch9329.ModLeftShift, "z") //stay
										time.Sleep(time.Millisecond * 50)
										for _, bound := range bounds.Boxes {
											_, _ = controlCl.MouseActionAbsolute(ch9329.MousePressLeft, image.Point{
												X: int((bound[2]-bound[0])/2) + bound[0],
												Y: bound[1] + 30,
											}, 0)
											time.Sleep(time.Millisecond * 40)
											_, _ = controlCl.MouseAbsoluteEnd()
											time.Sleep(time.Millisecond * 40)
											if val, ok := service.PlayerStats.Player[pid]; ok {
												playerStat = &val
											}
											if len(pidsStack[pid].preferredTargets) > 0 {
												if currentTarget, _ := service.GetCurrentTarget(logger); currentTarget != "" {
													logger.Info("prefered target is " + currentTarget)
													if !bestMatch(currentTarget, pidsStack[pid].allowedTargets, logger) {
														_, _ = controlCl.EndKey()
														time.Sleep(time.Millisecond * 50)
														_, _ = controlCl.SendKey(0, "esc")
														time.Sleep(time.Millisecond * 50)
														_, _ = controlCl.EndKey()
														time.Sleep(time.Millisecond * 50)
														_, _ = controlCl.SendKey(ch9329.ModLeftShift, "z") //stay
														i++
														continue
													}
												}
											}
											if playerStat.Target.HpPercent > 0 {
												break
											}
										}
										_, _ = controlCl.EndKey()
										//if playerStat.Target.HpPercent == 0 {
										//	controlCl.MouseActionAbsolute(ch9329.MousePressRight, image.Pt(0, 0), 0)
										//	time.Sleep(time.Millisecond * 200)
										//	controlCl.MouseActionAbsolute(ch9329.MousePressRight, image.Pt(5, 0), 0)
										//	time.Sleep(time.Millisecond * 100)
										//	controlCl.MouseActionAbsolute(0, image.Pt(10, 0), 0)
										//}
									}
									runAction.LastRun = time.Now()
								}
								i++
								time.Sleep(time.Millisecond * 1)
								continue
							}
							if runAction.Action == service.ActionAssistPartyMember {
								checksPassed, windowSwitched, i = handleAssistPartyMember(checksPassed, pid, controlCl, logger, runAction, windowSwitched, anotherPid, i)
								continue
							}

							checksPassed = makeChecks(pidsStack, pid, checksPassed, controlCl, logger)
							if !checksPassed {
								//logger.Error("makecheck failed")
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
							}

							if runAction.Action == service.ActionAttack {
								if playerStat.Target.HpPercent == 0 {
									i++
									continue
								}
								if len(pidsStack[pid].allowedTargets) > 0 {
									if currentTarget, _ := service.GetCurrentTarget(logger); currentTarget != "" {
										logger.Info("target is " + currentTarget)
										if !bestMatch(currentTarget, pidsStack[pid].allowedTargets, logger) {
											i++
											continue
										}
									}
								}
							}
							if strings.Contains(runAction.Binding, "+") {
								pieces := strings.Split(runAction.Binding, "+")
								var modifier byte
								switch pieces[0] {
								case "ctrl":
									modifier = ch9329.ModLeftCtrl
								case "alt":
									modifier = ch9329.ModLeftAlt
								default:
									modifier = 0
								}
								_, _ = controlCl.SendKey(modifier, pieces[1])
							} else {
								_, _ = controlCl.SendKey(0, runAction.Binding)
							}
							time.Sleep(time.Millisecond * 50)
							_, _ = controlCl.EndKey()
							if runAction.DelayMilliseconds > 0 {
								time.Sleep(time.Millisecond * time.Duration(runAction.DelayMilliseconds))
							}
							logger.Info("press ", runAction.Action, " ", runAction.Binding)
							if runAction.Action == service.ActionUnstuck {
								windowSwitched = handleUnstuck(checksPassed, pid, controlCl, logger, windowSwitched, anotherPid, runAction)
							}
							runAction.LastRun = time.Now()
							i++
							time.Sleep(time.Millisecond * time.Duration(randNum(5, 20)))
						}
						if windowSwitched {
							windowSwitched = false
							_ = switchWindow(anotherPid, controlCl, logger)
							if !pidsStack[anotherPid].TryLock() {
								pidsStack[anotherPid].waitCh <- struct{}{}
							} else {
								pidsStack[anotherPid].Unlock()
							}
						}
					}
				}
			}
		}()
	}
}

func handleAssistPartyMember(checksPassed bool, pid uint32, controlCl *service.Control, logger *zap.SugaredLogger, runAction *preset.Item, windowSwitched bool, anotherPid uint32, i int) (bool, bool, int) {
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
			_, _ = controlCl.MouseActionAbsolute(ch9329.MousePressRight, point, 0)
			_, _ = controlCl.MouseAbsoluteEnd()
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
	return checksPassed, windowSwitched, i
}

func handleUnstuck(checksPassed bool, pid uint32, controlCl *service.Control, logger *zap.SugaredLogger, windowSwitched bool, anotherPid uint32, runAction *preset.Item) bool {
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
		_, _ = controlCl.MouseActionAbsolute(ch9329.MousePressLeft, image.Point{960, 540 + 300}, 0)
		time.Sleep(time.Millisecond * 50)
		_, _ = controlCl.MouseAbsoluteEnd()
		time.Sleep(time.Second * 3)
		_, _ = controlCl.SendKey(0, runAction.Binding)
		time.Sleep(time.Millisecond * 50)
		_, _ = controlCl.EndKey()
		time.Sleep(time.Millisecond * 50)
		_, _ = controlCl.SendKey(0, "esc")
		time.Sleep(time.Millisecond * 50)
		_, _ = controlCl.EndKey()
		if runAction.DelayMilliseconds > 0 {
			time.Sleep(time.Millisecond * time.Duration(runAction.DelayMilliseconds))
		}
	}
	return windowSwitched
}

func initPartyMemberMap(cnf *core.Config) {
	for idx, val := range cnf.AssistPartyMemberMap {
		service.AssistPartyMemberMap[idx] = image.Point{val[0], val[1]}
	}
}

func bestMatch(input string, candidates []string, logger *zap.SugaredLogger) bool {
	best := ""
	bestDistance := int(^uint(0) >> 1)
	for _, candidate := range candidates {
		distance := levenshtein.ComputeDistance(
			input,
			candidate,
		)
		if distance <= TargetNameThreshold {
			return true
		}
		if distance < bestDistance {
			bestDistance = distance
			best = candidate
		}
	}
	logger.Info(input, " ", best, bestDistance)

	return false
}
