package core

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type logInfo struct {
	traceLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	logLevel      int
	rollFile      *lumberjack.Logger
}

func (logger *logInfo) Clear() {
	if logger.rollFile != nil {
		_ = logger.rollFile.Close()
	}
}

func (logger *logInfo) addMessageAndCodeInfo(fn string, line int, message string) string {
	now := time.Now().Format("15:04:05.000")
	return fmt.Sprintf("%s %s:%d: %s", now, filepath.Base(fn), line, message)
}

func (logger *logInfo) addMessage(message string) string {
	now := time.Now().Format("15:04:05.000")
	return fmt.Sprintf("%s: %s", now, message)
}

func (logger *logInfo) Trace(message string) {
	if !logger.IsTraceEnabled() {
		return
	}
	logger.traceLogger.Println(logger.addMessage(message))
}

func (logger *logInfo) IsTraceEnabled() bool {
	return logger.logLevel >= 3 && logger.traceLogger != nil
}

func (logger *logInfo) Info(message string) {
	logger.infoLogger.Println(logger.addMessage(message))
}

func (logger *logInfo) Error(message string) {
	_, fn, line, _ := runtime.Caller(1)
	logger.errorLogger.Println(logger.addMessageAndCodeInfo(fn, line, message))
}

func (logger *logInfo) Warning(message string) {
	logger.warningLogger.Println(logger.addMessage(message))
}

func (logger *logInfo) FatalError(message string) {
	_, fn, line, _ := runtime.Caller(1)
	logger.errorLogger.Fatalln(logger.addMessageAndCodeInfo(fn, line, message))
}

const (
	TracePrefix   = "TRACE:   "
	InfoPrefix    = "INFO:    "
	ErrorPrefix   = "ERROR:   "
	WarningPrefix = "WARNING: "
)

func InitDefaultLogging(useTrace bool) Logger {
	var result = logInfo{}

	options := log.Ldate

	if useTrace {
		result.logLevel = 3
		result.traceLogger = log.New(os.Stdout, TracePrefix, options)
	} else {
		result.logLevel = 2
	}

	result.infoLogger = log.New(os.Stdout, InfoPrefix, options)

	result.warningLogger = log.New(os.Stdout, WarningPrefix, options)

	result.errorLogger = log.New(os.Stderr, ErrorPrefix, options)

	return &result
}

func InitRollFileLogging(fileName string, useTrace bool) (Logger, error) {
	// Проверяем что файл для лога можно открыть или создать
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, err
	}

	file.Close()

	var rollLogger = &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    5, // megabytes
		MaxBackups: 500,
		MaxAge:     14,   //days
		Compress:   true, // disabled by default
	}

	var result = logInfo{
		rollFile: rollLogger,
	}

	options := log.Ldate

	if useTrace {
		result.logLevel = 3
		result.traceLogger = log.New(rollLogger, TracePrefix, options)
	} else {
		result.logLevel = 2
	}

	result.infoLogger = log.New(rollLogger, InfoPrefix, options)
	result.warningLogger = log.New(rollLogger, WarningPrefix, options)
	result.errorLogger = log.New(rollLogger, ErrorPrefix, options)

	return &result, nil
}
