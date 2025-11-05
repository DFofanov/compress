package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// FileLogger реализация логгера в файл
type FileLogger struct {
	file     *os.File
	logger   *log.Logger
	logLevel string
}

// NewFileLogger создает новый файловый логгер
func NewFileLogger(filename, logLevel string, maxSizeMB int, logToFile bool) (*FileLogger, error) {
	if !logToFile {
		return nil, nil
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	logger := log.New(file, "", log.LstdFlags)

	return &FileLogger{
		file:     file,
		logger:   logger,
		logLevel: strings.ToLower(logLevel),
	}, nil
}

// Debug логирует отладочное сообщение
func (l *FileLogger) Debug(format string, args ...interface{}) {
	if l.shouldLog("debug") {
		l.writeLog("DEBUG", format, args...)
	}
}

// Info логирует информационное сообщение
func (l *FileLogger) Info(format string, args ...interface{}) {
	if l.shouldLog("info") {
		l.writeLog("INFO", format, args...)
	}
}

// Warning логирует предупреждение
func (l *FileLogger) Warning(format string, args ...interface{}) {
	if l.shouldLog("warning") {
		l.writeLog("WARNING", format, args...)
	}
}

// Error логирует ошибку
func (l *FileLogger) Error(format string, args ...interface{}) {
	if l.shouldLog("error") {
		l.writeLog("ERROR", format, args...)
	}
}

// Success логирует успешное выполнение
func (l *FileLogger) Success(format string, args ...interface{}) {
	if l.shouldLog("info") {
		l.writeLog("SUCCESS", format, args...)
	}
}

// Close закрывает логгер
func (l *FileLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// writeLog записывает лог
func (l *FileLogger) writeLog(level, format string, args ...interface{}) {
	if l.logger == nil {
		return
	}

	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] %s", level, message)
}

// shouldLog проверяет, нужно ли логировать на данном уровне
func (l *FileLogger) shouldLog(level string) bool {
	levels := map[string]int{
		"debug":   0,
		"info":    1,
		"warning": 2,
		"error":   3,
	}

	currentLevel, ok := levels[l.logLevel]
	if !ok {
		currentLevel = 1 // default to info
	}

	messageLevel, ok := levels[level]
	if !ok {
		return false
	}

	return messageLevel >= currentLevel
}
