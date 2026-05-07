package logger

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity of a log message.
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	currentLevel = LevelInfo
	logFile      *os.File
	useColor     = true
)

// Init configures the logger. Pass an empty filepath to log to stdout only.
func Init(level LogLevel, filepath string) error {
	currentLevel = level
	useColor = filepath == "" // color only for terminal output

	if filepath != "" {
		f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		logFile = f
	}
	return nil
}

// Close flushes and closes the log file if one was configured.
func Close() {
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
}

func levelName(l LogLevel) string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO "
	case LevelWarn:
		return "WARN "
	case LevelError:
		return "ERROR"
	}
	return "INFO "
}

func levelColor(l LogLevel) string {
	if !useColor {
		return ""
	}
	switch l {
	case LevelDebug:
		return "\033[37m" // grey
	case LevelInfo:
		return "\033[36m" // cyan
	case LevelWarn:
		return "\033[33m" // yellow
	case LevelError:
		return "\033[31m" // red
	}
	return ""
}

func logMessage(level LogLevel, msg string) {
	if level < currentLevel {
		return
	}

	ts := time.Now().Format("15:04:05")
	reset := ""
	color := ""
	if useColor {
		color = levelColor(level)
		reset = "\033[0m"
	}

	line := fmt.Sprintf("%s%s %s%s %s", color, ts, levelName(level), reset, msg)
	fmt.Println(line)

	if logFile != nil {
		plain := fmt.Sprintf("%s %s %s", ts, levelName(level), msg)
		fmt.Fprintln(logFile, plain)
	}
}

// Debug logs a debug-level message (only shown when level is LevelDebug).
func Debug(format string, args ...interface{}) {
	if len(args) > 0 {
		logMessage(LevelDebug, fmt.Sprintf(format, args...))
	} else {
		logMessage(LevelDebug, format)
	}
}

// Info logs an informational message.
func Info(format string, args ...interface{}) {
	if len(args) > 0 {
		logMessage(LevelInfo, fmt.Sprintf(format, args...))
	} else {
		logMessage(LevelInfo, format)
	}
}

// Warn logs a warning message.
func Warn(format string, args ...interface{}) {
	if len(args) > 0 {
		logMessage(LevelWarn, fmt.Sprintf(format, args...))
	} else {
		logMessage(LevelWarn, format)
	}
}

// Error logs an error message.
func Error(format string, args ...interface{}) {
	if len(args) > 0 {
		logMessage(LevelError, fmt.Sprintf(format, args...))
	} else {
		logMessage(LevelError, format)
	}
}

// LevelNames returns all level names for display.
func LevelNames() []string {
	return []string{"DEBUG", "INFO", "WARN", "ERROR"}
}

// ParseLevel parses a level string into a LogLevel.
func ParseLevel(s string) (LogLevel, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARN", "WARNING":
		return LevelWarn, nil
	case "ERROR":
		return LevelError, nil
	}
	return LevelInfo, fmt.Errorf("unknown log level %q", s)
}
