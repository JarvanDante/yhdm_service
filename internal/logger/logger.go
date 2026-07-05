package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New 根据 level/format 构造一个 zap.Logger。
func New(level, format string) (*zap.Logger, error) {
	lvl := zap.NewAtomicLevel()
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl.SetLevel(zapcore.InfoLevel)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if format == "console" {
		cfg.Encoding = "console"
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return cfg.Build()
}
