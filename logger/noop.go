package logger

import "context"

var NoopLogger Logger = &noopLogger{}

type noopLogger struct{}

func (n noopLogger) Ctx(_ context.Context) Logger {
	return n
}

func (n noopLogger) With(_ ...Field) Logger {
	return n
}

func (n noopLogger) Debug(_ string, _ ...Field) {
}

func (n noopLogger) Info(_ string, _ ...Field) {
}

func (n noopLogger) Warn(_ string, _ ...Field) {
}

func (n noopLogger) Error(_ string, _ ...Field) {
}

func (n noopLogger) Panic(_ string, _ ...Field) {
}

func (n noopLogger) Fatal(_ string, _ ...Field) {
}
