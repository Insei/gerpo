package logger

import "context"

type Field interface {
	GetKey() string
	GetType() ValueType
	GetValue() any
}

type ValueType uint8

const (
	StringType ValueType = 15
)

type field struct {
	key       string
	value     string
	valueType ValueType
}

func (s field) GetKey() string {
	return s.key
}

func (s field) GetValue() any {
	return s.value
}

func (s field) GetType() ValueType {
	return s.valueType
}

func String(key, val string) Field {
	return field{
		key:       key,
		value:     val,
		valueType: StringType,
	}
}

type Logger interface {
	Ctx(ctx context.Context) Logger
	With(fields ...Field) Logger
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Panic(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}
