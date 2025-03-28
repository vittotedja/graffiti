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

func (l *Logger) Info(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *Logger) Error(msg string, err error) {
	l.entry.WithError(err).Error(msg)
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
