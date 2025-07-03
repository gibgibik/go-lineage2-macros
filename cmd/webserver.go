package cmd

import (
	"fmt"
	"time"

	"github.com/gibgibik/go-lineage2-macros/internal/core"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func createWebServerCommand(logger *zap.Logger) *cobra.Command {
	var webServer = &cobra.Command{
		Use: "web-server",
		Run: func(cmd *cobra.Command, args []string) {
			cnf := cmd.Context().Value("cnf").(*core.Config)
			fmt.Println(cnf)
			for {
				select {
				case <-cmd.Context().Done():
					logger.Info("web-server has stopped properly")
					return
				default:
					logger.Info("tick")
					time.Sleep(time.Second)
				}
			}
		},
	}
	return webServer
}
