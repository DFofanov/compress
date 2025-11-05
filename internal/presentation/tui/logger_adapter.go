package tui

import (
	"compressor/internal/domain/repositories"
	"fmt"
)

// UILogger адаптер логгера для отображения в UI
type UILogger struct {
	fileLogger repositories.Logger
	tuiManager *Manager
}

// NewUILogger создает новый UI логгер
func NewUILogger(fileLogger repositories.Logger, tuiManager *Manager) *UILogger {
	return &UILogger{
		fileLogger: fileLogger,
		tuiManager: tuiManager,
	}
}

// Debug логирует отладочное сообщение
func (l *UILogger) Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.fileLogger != nil {
		l.fileLogger.Debug(format, args...)
	}
	if l.tuiManager != nil {
		l.tuiManager.AddLog("DEBUG", message)
	}
}

// Info логирует информационное сообщение
func (l *UILogger) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.fileLogger != nil {
		l.fileLogger.Info(format, args...)
	}
	if l.tuiManager != nil {
		l.tuiManager.AddLog("INFO", message)
	}
}

// Warning логирует предупреждение
func (l *UILogger) Warning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.fileLogger != nil {
		l.fileLogger.Warning(format, args...)
	}
	if l.tuiManager != nil {
		l.tuiManager.AddLog("WARNING", message)
	}
}

// Error логирует ошибку
func (l *UILogger) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.fileLogger != nil {
		l.fileLogger.Error(format, args...)
	}
	if l.tuiManager != nil {
		l.tuiManager.AddLog("ERROR", message)
	}
}

// Success логирует успешное выполнение
func (l *UILogger) Success(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.fileLogger != nil {
		l.fileLogger.Success(format, args...)
	}
	if l.tuiManager != nil {
		l.tuiManager.AddLog("SUCCESS", message)
	}
}

// Close закрывает логгер
func (l *UILogger) Close() error {
	if l.fileLogger != nil {
		return l.fileLogger.Close()
	}
	return nil
}
