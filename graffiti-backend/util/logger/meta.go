package logger

import (
	"context"
	"github.com/sirupsen/logrus"
	"os"
)

type ctxKey string

const metadataKey ctxKey = "log_meta"

type Meta struct {
	RequestID string
	Route     string
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

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (m *Meta) GetLogger() *Logger {
	entry := logrus.WithFields(logrus.Fields{
		"request_id": m.RequestID,
		"route":      m.Route,
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

func Setup() error {
	// Open log files
	infoFile, err := os.OpenFile("info.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	errorFile, err := os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Base logger config
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.DebugLevel)

	// Prevent logs from going to stdout
	logrus.SetOutput(os.Stdout)

	// Add hooks
	logrus.AddHook(&LevelHook{
		Writer: infoFile,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
	})

	logrus.AddHook(&LevelHook{
		Writer: errorFile,
		LogLevels: []logrus.Level{
			logrus.WarnLevel,
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		},
	})

	return nil
}
