package repositories

// Logger интерфейс для логирования
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Error(format string, args ...interface{})
	Success(format string, args ...interface{})
	Close() error
}
