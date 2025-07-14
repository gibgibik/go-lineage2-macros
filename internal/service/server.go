package service

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/gibgibik/go-lineage2-macros/internal/core/entity"
	"github.com/gibgibik/go-lineage2-macros/internal/core/http"
	"go.uber.org/zap"
)

var (
	PlayerStat           *entity.PlayerStat
	targetHpWasPresentAt time.Time
	httpCl               = http.NewHttpClient()
)

type BoundsResult struct {
	Boxes [][]int `json:"boxes"`
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
			PlayerStat.Target.HpWasPresentAt = targetHpWasPresentAt.Unix()
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
	bounds, err := httpCl.RawGet(url)
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
		if i+1 < len(boxes.Boxes) && (boxes.Boxes[i+1][0]-boxes.Boxes[i][2]) <= 3 { //glue nested boxes
			result.Boxes[i+1] = []int{boxes.Boxes[i][0], boxes.Boxes[i][1], boxes.Boxes[i+1][2], boxes.Boxes[i+1][3]}
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
