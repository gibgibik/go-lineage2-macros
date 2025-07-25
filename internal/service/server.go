package service

import (
	"encoding/json"
	"fmt"
	http2 "net/http"
	"sync"
	"time"

	"github.com/gibgibik/go-lineage2-macros/internal/core/http"
	"github.com/gibgibik/go-lineage2-server/pkg/entity"
	"go.uber.org/zap"
)

var (
	PlayerStats                entity.StatStr
	PlayerStatsMutex           sync.Mutex
	targetHpWasPresentAt       time.Time
	fullTargetHpUnchangedSince time.Time
)

type BoundsResult struct {
	TargetName string  `json:"target_name"`
	Boxes      [][]int `json:"boxes"`
}

type InitData struct {
	PidsData map[uint32]string
}

type ForeGroundWindowInfo struct {
	Pid uint32 `json:"pid"`
}

//func StartPlayerStatUpdate(ctx context.Context, logger *zap.SugaredLogger) {
//	var err error
//	logger.Debug("player stat update start")
//	for {
//		select {
//		case <-ctx.Done():
//			logger.Info("player stat update stopped")
//			return
//		default:
//			PlayerStats, err = http.HttpCl.Get("")
//			if PlayerStats == nil {
//				continue
//			}
//			if PlayerStats.Target.HpPercent > 0 {
//				targetHpWasPresentAt = time.Now()
//			}
//			if PlayerStats.Target.HpPercent >= 99 {
//				if fullTargetHpUnchangedSince.IsZero() {
//					fullTargetHpUnchangedSince = time.Now()
//				}
//			} else {
//				fullTargetHpUnchangedSince = time.Now()
//			}
//			PlayerStats.Target.HpWasPresentAt = targetHpWasPresentAt.Unix()
//			PlayerStats.Target.FullHpUnchangedSince = fullTargetHpUnchangedSince.Unix()
//			if err != nil {
//				logger.Error("player pull stat error: ", err.Error())
//				continue
//			}
//			time.Sleep(time.Millisecond * 100)
//		}
//	}
//}

func FindBounds(logger *zap.SugaredLogger) (*BoundsResult, error) {
	var err error
	logger.Info("get bounds start")
	bounds, err := http.HttpCl.RawRequest("findBounds", http2.MethodGet, nil)
	logger.Info("get bounds end")
	if err != nil {
		return nil, err
	}
	var boxes BoundsResult
	err = json.Unmarshal(bounds, &boxes)
	if err != nil {
		logger.Error("parse bounds json error: ", err.Error())
		return nil, nil
	}
	//if len(boxes.Boxes) > 10 {
	//	boxes.Boxes = boxes.Boxes[:10]
	//}
	return &boxes, nil
}

func Init() (InitData, error) {
	var result InitData
	initData, err := http.HttpCl.RawRequest("init", http2.MethodGet, nil)
	if err != nil {
		return InitData{}, err
	}
	_ = json.Unmarshal(initData, &result)
	return result, nil
}
func GetForegroundWindowPid() (uint32, error) {
	res, err := http.HttpCl.RawRequest("getForegroundWindowPid", http2.MethodPost, nil)
	fmt.Println("get foreground", string(res))
	if err != nil {
		fmt.Println("err", err)
		return 0, err
	}
	var result ForeGroundWindowInfo
	_ = json.Unmarshal(res, &result)
	return result.Pid, nil
}
