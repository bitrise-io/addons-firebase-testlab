package logger

import (
	"context"
	"fmt"

	"github.com/gobuffalo/buffalo"
	"go.uber.org/zap"
)

type ctxKeyType string

const loggerKey ctxKeyType = "ctx-logger"

var logger *zap.Logger

func init() {
	newLogger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %s", err)
	}
	logger = newLogger
}

// NewContext ...
func NewContext(ctx buffalo.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, loggerKey, WithContext(ctx).With(fields...))
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
