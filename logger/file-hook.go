package logger

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	logrus "github.com/sirupsen/logrus"
)

type FileHook struct {
	Writer          io.Writer
	TimestampFormat string
	LogLevels       []logrus.Level
}

func NewFileHook(writer io.Writer, timestampFormat string, level logrus.Level) *FileHook {
	return &FileHook{
		Writer:          writer,
		TimestampFormat: timestampFormat,
		LogLevels:       getLogLevels(level),
	}
}

func (hook *FileHook) Fire(entry *logrus.Entry) error {
	// Uppercase the level text and truncate it to 4 chars
	levelText := strings.ToUpper(entry.Level.String())
	levelText = levelText[0:4]

	// Extract and sort all the keys alphabetically
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Print log entry to a buffer and append a newline
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "%s[%s] %-64s ", levelText, entry.Time.Format(hook.TimestampFormat), entry.Message)
	for _, k := range keys {
		v := entry.Data[k]
		fmt.Fprintf(b, " %s=", k)
		appendValue(b, v)
	}
	b.WriteByte('\n')

	// Write buffer to file
	_, err := hook.Writer.Write(b.Bytes())
	return err
}

func (hook *FileHook) Levels() []logrus.Level {
	return hook.LogLevels
}

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
