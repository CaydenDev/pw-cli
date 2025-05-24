package util

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

type Logger struct {
	mu       sync.Mutex
	file     *os.File
	minLevel LogLevel
	showTime bool
	showFile bool
	maxSize  int64
	maxFiles int
	filePath string
}

func NewLogger(logDir string, minLevel LogLevel) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, err
	}

	logPath := filepath.Join(logDir, "pwvault.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:     file,
		minLevel: minLevel,
		showTime: true,
		showFile: true,
		maxSize:  5 * 1024 * 1024, // 5MB
		maxFiles: 3,
		filePath: logPath,
	}, nil
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

func (l *Logger) SetShowTime(show bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.showTime = show
}

func (l *Logger) SetShowFile(show bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.showFile = show
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.minLevel {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.rotateIfNeeded(); err != nil {
		fmt.Fprintf(os.Stderr, "Error rotating log file: %v\n", err)
	}

	var builder strings.Builder

	if l.showTime {
		builder.WriteString(time.Now().Format("2006-01-02 15:04:05 "))
	}

	builder.WriteString(fmt.Sprintf("[%s] ", l.levelString(level)))

	if l.showFile {
		if file, line := l.getCallerInfo(); file != "" {
			builder.WriteString(fmt.Sprintf("%s:%d ", file, line))
		}
	}

	builder.WriteString(fmt.Sprintf(format, args...))
	builder.WriteString("\n")

	if _, err := l.file.WriteString(builder.String()); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to log file: %v\n", err)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warning(format string, args ...interface{}) {
	l.log(WARNING, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) levelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (l *Logger) getCallerInfo() (string, int) {
	if pc, file, line, ok := runtime.Caller(3); ok {
		if function := runtime.FuncForPC(pc); function != nil {
			file = filepath.Base(file)
			return file, line
		}
	}
	return "", 0
}

func (l *Logger) rotateIfNeeded() error {
	info, err := l.file.Stat()
	if err != nil {
		return err
	}

	if info.Size() < l.maxSize {
		return nil
	}

	l.file.Close()

	for i := l.maxFiles - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.filePath, i)
		newPath := fmt.Sprintf("%s.%d", l.filePath, i+1)

		if _, err := os.Stat(oldPath); err == nil {
			if err := os.Rename(oldPath, newPath); err != nil {
				return err
			}
		}
	}

	if err := os.Rename(l.filePath, l.filePath+".1"); err != nil {
		return err
	}

	file, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	l.file = file
	return nil
}
