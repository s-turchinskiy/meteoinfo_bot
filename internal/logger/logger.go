// Package logger Логирование для агента
package logger

import (
	"errors"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var errCannotInitializeZap = errors.New("cannot initialize zap")

func Initialize(paths []string) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = paths
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)

	logger, err := cfg.Build()
	if err != nil {
		return nil, errCannotInitializeZap
	}

	return logger.Sugar(), nil
}
