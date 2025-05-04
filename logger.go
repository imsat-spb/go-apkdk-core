package core

type Logger interface {
	Clear()
	Info(message string)
	Error(message string)
	Warning(message string)
	Trace(message string)
	IsTraceEnabled() bool
	FatalError(message string)
}
