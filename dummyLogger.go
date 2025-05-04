package core

type DummyLogger struct {
}

func (*DummyLogger) Clear() {

}

func (*DummyLogger) Info(message string) {

}

func (*DummyLogger) Error(message string) {

}

func (*DummyLogger) Warning(message string) {

}
func (*DummyLogger) Trace(message string) {

}

func (*DummyLogger) FatalError(message string) {

}

func (*DummyLogger) IsTraceEnabled() bool {
	return false
}
