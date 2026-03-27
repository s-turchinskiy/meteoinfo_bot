// Package logger Логирование для агента
package logger

import (
	"errors"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log                    = zap.NewNop().Sugar()
	ErrCannotInitializeZap = errors.New("cannot initialize zap")
)

func Initialize() error {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"bot.log", "stdout"}
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)

	logger, err := cfg.Build()
	if err != nil {
		return ErrCannotInitializeZap
	}

	Log = logger.Sugar()

	return nil
}
