package logger

import (
	"bytes"
	"fmt"

	logrus "github.com/sirupsen/logrus"
)

func appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}
	if !needsQuoting(stringVal) {
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}

func needsQuoting(text string) bool {
	if len(text) == 0 {
		return true
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}

func getLogLevels(level logrus.Level) []logrus.Level {
	switch level {
	case logrus.DebugLevel:
		return []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
			logrus.DebugLevel,
		}
	case logrus.InfoLevel:
		return []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
		}
	case logrus.WarnLevel:
		return []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		}
	case logrus.ErrorLevel:
		return []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		}
	case logrus.FatalLevel:
		return []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
		}
	default:
		return []logrus.Level{
			logrus.PanicLevel,
		}
	}
}
