package utils

import (
"log"
"os"
"strings"
)

type LogLevel int

const (
DEBUG LogLevel = iota
INFO
WARN
ERROR
)

type Logger struct {
	level LogLevel
}

var defaultLogger *Logger

func InitLogger(level string) {
	logLevel := parseLogLevel(level)
	defaultLogger = &Logger{level: logLevel}
}

func parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

func Debug(message string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= DEBUG {
		log.Printf("[DEBUG] "+message, args...)
	}
}

func Info(message string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= INFO {
		log.Printf("[INFO] "+message, args...)
	}
}

func Warn(message string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= WARN {
		log.Printf("[WARN] "+message, args...)
	}
}

func Error(message string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= ERROR {
		log.Printf("[ERROR] "+message, args...)
	}
}

func Fatal(message string, args ...interface{}) {
	log.Printf("[FATAL] "+message, args...)
	os.Exit(1)
}

func GetLogger() *Logger {
	if defaultLogger == nil {
		InitLogger("info")
	}
	return defaultLogger
}
