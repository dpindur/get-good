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
