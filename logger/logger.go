package logger

import (
	"os"

	logrus "github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func ConfigureLogger(level logrus.Level, file *os.File) {
	timestampFormat := "2006/01/15 15:04:05"
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: timestampFormat,
	}
	hook := NewFileHook(file, timestampFormat, level)
	Logger.Formatter = formatter
	Logger.Level = level
	Logger.Out = os.Stdout
	Logger.AddHook(hook)
}
