package logger

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	logrus "github.com/sirupsen/logrus"
)

type Terminal interface {
	AddLog(log []byte)
}

type TerminalHook struct {
	Terminal        Terminal
	TimestampFormat string
	LogLevels       []logrus.Level
}

func NewTerminalHook(terminal Terminal, timestampFormat string, level logrus.Level) *TerminalHook {
	return &TerminalHook{
		Terminal:        terminal,
		TimestampFormat: timestampFormat,
		LogLevels:       getLogLevels(level),
	}
}

func (hook *TerminalHook) Fire(entry *logrus.Entry) error {
	// Uppercase the level text and truncate it to 4 chars
	levelText := strings.ToUpper(entry.Level.String())
	levelText = levelText[0:4]

	// Extract and sort all the keys alphabetically
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Print log entry to a buffer
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "%s[%s] %-64s ", levelText, entry.Time.Format(hook.TimestampFormat), entry.Message)
	for _, k := range keys {
		v := entry.Data[k]
		fmt.Fprintf(b, " %s=", k)
		appendValue(b, v)
	}

	// Write buffer to terminal
	hook.Terminal.AddLog(b.Bytes())
	return nil
}

func (hook *TerminalHook) Levels() []logrus.Level {
	return hook.LogLevels
}
