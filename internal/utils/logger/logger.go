package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	LevelTrace = 0
	LevelDebug = 1
	LevelInfo  = 2
	LevelWarn  = 3
	LevelError = 4
)

var levelStrings = map[int]string{
	LevelTrace: "TRACE",
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

var levelNumbers = map[string]int{
	"TRACE": LevelTrace,
	"DEBUG": LevelDebug,
	"INFO":  LevelInfo,
	"WARN":  LevelWarn,
	"ERROR": LevelError,
}

type LogEntry struct {
	Timestamp   string `json:"timestamp"`
	Level       string `json:"level"`
	ServiceName string `json:"service_name"`
	Message     string `json:"message"`
}

type Logger struct {
	mutex      sync.RWMutex
	service    string
	level      int
	out        io.Writer
	timeFormat string
}

func New(serviceName, level string) *Logger {
	l := &Logger{
		service:    serviceName,
		level:      LevelInfo,
		out:        os.Stdout,
		timeFormat: time.RFC3339,
	}

	if level != "" {
		l.SetLevel(level)
	}

	return l
}

func (lgr *Logger) SetOutput(w io.Writer) {
	lgr.mutex.Lock()
	lgr.out = w
	lgr.mutex.Unlock()
}

func (lgr *Logger) SetLevel(level string) {
	lgr.mutex.Lock()
	defer lgr.mutex.Unlock()
	levelUpper := strings.ToUpper(level)
	if num, ok := levelNumbers[levelUpper]; ok {
		lgr.level = num
	}
}

func (lgr *Logger) Trace(message string) {
	lgr.log(LevelTrace, message)
}

func (lgr *Logger) Debug(message string) {
	lgr.log(LevelDebug, message)
}

func (lgr *Logger) Info(message string) {
	lgr.log(LevelInfo, message)
}

func (lgr *Logger) Warn(message string) {
	lgr.log(LevelWarn, message)
}

func (lgr *Logger) Error(message string) {
	lgr.log(LevelError, message)
}

func (lgr *Logger) shouldLog(level int) bool {
	lgr.mutex.RLock()
	defer lgr.mutex.RUnlock()
	return level >= lgr.level
}

func (lgr *Logger) log(level int, message string) {
	if !lgr.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp:   time.Now().Format(lgr.timeFormat),
		Level:       levelStrings[level],
		ServiceName: lgr.service,
		Message:     message,
	}

	b, _ := json.Marshal(entry)

	lgr.mutex.RLock()
	out := lgr.out
	lgr.mutex.RUnlock()

	_, _ = fmt.Fprintln(out, string(b))
}
