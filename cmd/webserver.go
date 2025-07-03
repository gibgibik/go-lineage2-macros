package cmd

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gibgibik/go-lineage2-macros/internal/core"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func createWebServerCommand(logger *zap.SugaredLogger) *cobra.Command {
	var webServer = &cobra.Command{
		Use: "web-server",
		RunE: func(cmd *cobra.Command, args []string) error {
			startResult := make(chan error, 1)
			cnf := cmd.Context().Value("cnf").(*core.Config)
			handle := &http.Server{
				Addr:           ":" + cnf.WebServer.Port,
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
				ErrorLog:       log.New(&core.FwdToZapWriter{logger}, "", 0),
			}
			go func() {
				if err := handle.ListenAndServe(); err != nil {
					if !errors.Is(err, http.ErrServerClosed) {
						logger.Error(err)
					}
					startResult <- err
				}
			}()
			fmt.Println("ok")
			for {
				select {
				case <-cmd.Context().Done():
					err := handle.Shutdown(cmd.Context())
					logger.Info(fmt.Sprintf("web-server stop result: %v", err))
					return err
				case <-startResult:
					fmt.Println("end received")
					return nil
				default:
					fmt.Println("tick")
					time.Sleep(time.Microsecond * 100000)
				}
			}
		},
	}
	return webServer
}
