package logger

import (
	"io/ioutil"
	"os"

	logrus "github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func ConfigureLogger(level logrus.Level, terminal Terminal, file *os.File) {
	timestampFormat := "2006/01/15 15:04:05"
	fileHook := NewFileHook(file, timestampFormat, level)
	terminalHook := NewTerminalHook(terminal, timestampFormat, level)
	Logger.Level = level
	Logger.Out = ioutil.Discard
	Logger.AddHook(fileHook)
	Logger.AddHook(terminalHook)
}
