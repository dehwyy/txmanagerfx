package txcore

type Logger interface {
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type nopLogger struct{}

func (nopLogger) Warn(string, ...any) {}

func (nopLogger) Error(string, ...any) {}

func NopLogger() Logger {
	return nopLogger{}
}
