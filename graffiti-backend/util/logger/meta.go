package logger

import (
	"context"
	"github.com/sirupsen/logrus"
)

type ctxKey string

const metadataKey ctxKey = "log_meta"

type Meta struct {
	RequestID string
}

type Logger struct {
	entry *logrus.Entry
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

// Error logs an error message with an actual error
func (l *Logger) Error(msg string, err error) {
	l.entry.WithError(err).Error(msg)
}

// Errorf logs a business error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (m *Meta) GetLogger() *Logger {
	entry := logrus.WithFields(logrus.Fields{
		"request_id": m.RequestID,
	})
	return &Logger{entry: entry}
}

func GetMetadata(ctx context.Context) *Meta {
	meta, ok := ctx.Value(metadataKey).(*Meta)
	if !ok {
		return &Meta{}
	}
	return meta
}
