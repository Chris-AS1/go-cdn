package logger

import (
	"fmt"
	"go-cdn/config"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger(cfg *config.Config) *zap.SugaredLogger {
	writer_file := zapcore.AddSync(&lumberjack.Logger{
		Filename: fmt.Sprintf("%s/go-cdn-%s.log",
			strings.TrimRight(cfg.Telemetry.LogPath, "/"),
			cfg.Consul.ConsulServiceID),
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	writer_console := zapcore.AddSync(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			writer_file,
			zap.InfoLevel,
		),
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			writer_console,
			zap.DebugLevel,
		),
	)
	logger := zap.New(core, zap.AddCaller())
	sugar := logger.Sugar()
	return sugar
}
