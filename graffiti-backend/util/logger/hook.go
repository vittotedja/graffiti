package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

type LevelHook struct {
	Writer    *os.File
	LogLevels []logrus.Level
}

func (hook *LevelHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

func (hook *LevelHook) Levels() []logrus.Level {
	return hook.LogLevels
}
