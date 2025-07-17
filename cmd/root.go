package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gibgibik/go-lineage2-macros/internal/core"
	"github.com/gibgibik/go-lineage2-macros/internal/core/http"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var rootCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start")
	},
}

func Execute() error {
	var err error
	pe := zap.NewProductionEncoderConfig()
	pe.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(pe)

	wsEncoded := zap.NewDevelopmentEncoderConfig()
	wsEncoded.EncodeTime = zapcore.RFC3339TimeEncoder

	webEncoder := zap.NewDevelopmentEncoderConfig()
	webEncoder.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		l.Set("")
	}
	webEncoder.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")

	f, err := os.OpenFile(fmt.Sprintf("var/log/app.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	w := zapcore.AddSync(f)
	cZ := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(pe),
			w,
			zap.InfoLevel,
		),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
		zapcore.NewCore(zapcore.NewConsoleEncoder(webEncoder), zapcore.AddSync(BaseWsSender{}), zapcore.InfoLevel),
	)
	logger := zap.New(cZ)
	cnf, err := core.InitConfig()
	if err != nil {
		return err
	}
	http.IniHttpClient(cnf.BaseUrl)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	defer cancel()
	rootCmd := &cobra.Command{
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	rootCmd.AddCommand(createWebServerCommand(logger.Sugar()))
	go func() {
		defer cancel()
		err = rootCmd.ExecuteContext(context.WithValue(ctx, "cnf", cnf))
	}()
	<-ctx.Done()
	logger.Info("shutdown start")
	time.Sleep(time.Second * 5)
	logger.Info("shutdown end")
	return err
}
