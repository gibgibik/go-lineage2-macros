package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	http2 "net/http"
	"sort"
	"time"

	"github.com/gibgibik/go-lineage2-macros/internal/core/entity"
	"github.com/gibgibik/go-lineage2-macros/internal/core/http"
	"go.uber.org/zap"
)

var (
	PlayerStat                 *entity.PlayerStat
	targetHpWasPresentAt       time.Time
	fullTargetHpUnchangedSince time.Time
	httpCl                     = http.NewHttpClient()
)

type BoundsResult struct {
	Boxes [][]int `json:"boxes"`
}

type InitData struct {
	PidsData map[uint32]string
}

type ChangeCurrentWindowStr struct {
	Pid uint32 `json:"pid"`
}

func StartPlayerStatUpdate(ctx context.Context, url string, logger *zap.SugaredLogger) error {
	var err error
	logger.Debug("player stat update start")
	for {
		select {
		case <-ctx.Done():
			logger.Info("player stat update stopped")
			return nil
		default:
			PlayerStat, err = httpCl.Get(url)
			if PlayerStat == nil {
				continue
			}
			if PlayerStat.Target.HpPercent > 0 {
				targetHpWasPresentAt = time.Now()
			}
			if PlayerStat.Target.HpPercent >= 99 {
				if fullTargetHpUnchangedSince.IsZero() {
					fullTargetHpUnchangedSince = time.Now()
				}
			} else {
				fullTargetHpUnchangedSince = time.Now()
			}
			PlayerStat.Target.HpWasPresentAt = targetHpWasPresentAt.Unix()
			PlayerStat.Target.FullHpUnchangedSince = fullTargetHpUnchangedSince.Unix()
			if err != nil {
				logger.Error("player pull stat error: ", err.Error())
				continue
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func FindBounds(url string, logger *zap.SugaredLogger) ([][]int, error) {
	var err error
	bounds, err := httpCl.RawRequest(url, http2.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	var boxes BoundsResult
	var result BoundsResult
	err = json.Unmarshal(bounds, &boxes)
	if err != nil {
		logger.Error("parse bounds json error: ", err.Error())
		return nil, nil
	}
	sort.Slice(boxes.Boxes, func(i, j int) bool {
		return boxes.Boxes[i][1] < boxes.Boxes[j][1]
	})
	sort.Slice(boxes.Boxes, func(i, j int) bool {
		return boxes.Boxes[i][0] < boxes.Boxes[j][2]
	})
	for i := 0; i < len(boxes.Boxes); i++ {
		if i+1 < len(boxes.Boxes) && (boxes.Boxes[i+1][0]-boxes.Boxes[i][2]) <= 3 && math.Abs(float64(boxes.Boxes[i+1][1]-boxes.Boxes[i][3])) < 3 { //glue nested boxes
			result.Boxes = append(result.Boxes, []int{boxes.Boxes[i][0], boxes.Boxes[i][1], boxes.Boxes[i+1][2], boxes.Boxes[i+1][3]})
			i++
		} else {
			result.Boxes = append(result.Boxes, boxes.Boxes[i])
		}
	}
	//if len(boxes.Boxes) > 10 {
	//	boxes.Boxes = boxes.Boxes[:10]
	//}
	return result.Boxes, nil
}

func Init(url string, logger *zap.SugaredLogger) (InitData, error) {
	var result InitData
	initData, err := httpCl.RawRequest(url, http2.MethodGet, nil)
	if err != nil {
		return InitData{}, err
	}
	_ = json.Unmarshal(initData, &result)
	return result, nil
}
func ChangeCurrentWindow(url string, pid *ChangeCurrentWindowStr, logger *zap.SugaredLogger) error {
	b, _ := json.Marshal(pid)
	fmt.Println(url)
	res, err := httpCl.RawRequest(url, http2.MethodPost, bytes.NewBuffer(b))
	if err != nil {
		logger.Error("change current window error: ", err.Error())
		return err
	}
	logger.Debug("change current window result ", string(res))
	return nil
}
