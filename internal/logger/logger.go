package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new SugaredLogger with the specified log level.
// level — a string representing the log level, e.g., "debug", "info", "warn", "error".
func New(level string) (*zap.SugaredLogger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
