package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gibgibik/go-lineage2-macros/internal/core/entity"
	"github.com/gibgibik/go-lineage2-macros/internal/core/http"
	"go.uber.org/zap"
)

var (
	PlayerStat *entity.PlayerStat
)

func StartPlayerStatUpdate(ctx context.Context, url string, logger *zap.SugaredLogger) error {
	var err error
	logger.Debug("player stat update start")
	httpCl := http.NewHttpClient()
	for {
		select {
		case <-ctx.Done():
			logger.Info("player stat update stopped")
			return nil
		default:
			PlayerStat, err = httpCl.Get(url)
			if PlayerStat.Target.HpPercent > 0 {
				fmt.Println("HP", PlayerStat.Target.HpPercent)
				PlayerStat.Target.HpWasPresentAt = time.Now().Unix()
			}
			if err != nil {
				logger.Error("player pull stat error: ", err.Error())
				continue
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}
