package logger

import (
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Logger ...
type Logger struct {
	logger *zap.Logger
}

// New ...
func New() (Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return Logger{}, errors.WithStack(err)
	}
	return Logger{logger: logger}, nil
}

// Sync ...
func (l *Logger) Sync() func() {
	return func() {
		err := l.logger.Sync()
		if err != nil {
			fmt.Printf("Failed to sync logger: %s", err)
		}
	}
}

// Error ...
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Warn ...
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Info ...
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}
