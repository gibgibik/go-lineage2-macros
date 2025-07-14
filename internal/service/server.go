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
	var result BoundsResult
	err = json.Unmarshal(bounds, &result)
	if err != nil {
		logger.Error("parse bounds json error: ", err.Error())
		return nil, nil
	}
	sort.Slice(result.Boxes, func(i, j int) bool {
		return result.Boxes[i][1] > result.Boxes[j][1]
	})
	//if len(result.Boxes) > 10 {
	//	result.Boxes = result.Boxes[:10]
	//}
	return result.Boxes, nil
}
