package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func createWebServerCommand(logger *zap.Logger) *cobra.Command {
	var webServer = &cobra.Command{
		Use: "web-server",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("test")
		},
	}
	return webServer
}
