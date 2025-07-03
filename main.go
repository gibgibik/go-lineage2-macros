package main

import (
	"fmt"
	"os"

	"github.com/gibgibik/go-lineage2-macros/cmd"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	pe := zap.NewProductionEncoderConfig()
	pe.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(pe)
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
	)
	logger := zap.New(cZ)
	logger.Info("start")
	cmd.Execute(logger)
}
