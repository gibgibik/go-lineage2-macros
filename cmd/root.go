package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gibgibik/go-lineage2-macros/cmd/web"
	"github.com/gibgibik/go-lineage2-macros/internal/core"
	"github.com/gibgibik/go-lineage2-macros/internal/core/http"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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

	f, err := os.OpenFile("var/log/app.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	// Wrap the file sink in a buffered syncer with a periodic flush so log
	// lines reach app.log even when the process is stuck or killed hard.
	// Without this, zap's internal buffer can hold log output indefinitely,
	// which is why the log appeared empty during hangs.
	fileWS := &zapcore.BufferedWriteSyncer{
		WS:            zapcore.AddSync(f),
		FlushInterval: time.Second,
	}
	defer fileWS.Stop() // flushes any remaining buffered output on exit
	w := zapcore.WriteSyncer(fileWS)
	cZ := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(pe),
			w,
			zap.InfoLevel,
		),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
		zapcore.NewCore(zapcore.NewConsoleEncoder(webEncoder), zapcore.AddSync(web.BaseWsSender{}), zapcore.InfoLevel),
	)
	logger := zap.New(cZ)
	defer logger.Sync() // ensure buffered log entries are flushed on shutdown
	rootCmd := &cobra.Command{
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	rootCmd.AddCommand(web.CreateWebServerCommand(logger.Sugar()))
	rootCmd.PersistentFlags().StringP("config", "c", "", "")
	cnf, err := core.InitConfig()
	if err != nil {
		return err
	}
	http.IniHttpClient(cnf.BaseUrl)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	go func() {
		defer cancel()
		err = rootCmd.ExecuteContext(context.WithValue(ctx, web.CtxKeyConfig, cnf))
	}()
	<-ctx.Done()
	logger.Info("shutdown start")
	time.Sleep(time.Second * 5)
	logger.Info("shutdown end")
	_ = logger.Sync()
	return err
}
