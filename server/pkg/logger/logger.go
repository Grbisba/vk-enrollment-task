package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func highPriorityLevelEnablerFunc() zap.LevelEnablerFunc {
	return func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	}
}

func lowPriorityLevelEnablerFunc() zap.LevelEnablerFunc {
	return func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	}
}

func NewProduction(opts ...zap.Option) (*zap.Logger, error) {
	return NewWithOpts(opts...)
}

// NewWithOpts collects zap.Field as zap.Option then call NewProduction with option
func NewWithOpts(opts ...zap.Option) (*zap.Logger, error) {
	highPriority := highPriorityLevelEnablerFunc()
	lowPriority := lowPriorityLevelEnablerFunc()

	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, consoleErrors, highPriority),
		zapcore.NewCore(jsonEncoder, consoleDebugging, lowPriority),
	)

	logger := zap.New(core, opts...)

	zap.ReplaceGlobals(logger)

	return logger, nil
}
