package logging

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const loggerKey string = "ctx-logger"

var logger *zap.Logger

func init() {
	newLogger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %s", err)
	}
	logger = newLogger
}

// NewContext ...
func NewContext(ctx buffalo.Context, fields ...zap.Field) buffalo.Context {
	ctx.Set(loggerKey, WithContext(ctx).With(fields...))
	return ctx
}

// WithContext ...
func WithContext(ctx buffalo.Context) *zap.Logger {
	if ctx == nil {
		return logger
	}
	if ctxLogger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return ctxLogger
	}
	return logger
}

// Sync ...
func Sync(logger *zap.Logger) {
	err := logger.Sync()
	if err != nil {
		if !strings.Contains(err.Error(), "invalid argument") {
			fmt.Printf("Failed to sync logger: %s", errors.WithStack(err))
		}
	}
}
