package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func createWebServerCommand(logger *zap.Logger) *cobra.Command {
	var webServer = &cobra.Command{
		Use: "web-server",
		Run: func(cmd *cobra.Command, args []string) {

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
