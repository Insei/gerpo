package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoopLogger_AllMethodsNoOp(t *testing.T) {
	ctx := context.Background()
	l := NoopLogger

	// None of these should panic nor return nil.
	assert.NotNil(t, l.Ctx(ctx))
	assert.NotNil(t, l.With(String("k", "v")))

	// The log-level methods are side-effect-free.
	l.Debug("msg")
	l.Info("msg", String("k", "v"))
	l.Warn("msg")
	l.Error("msg")
	l.Panic("msg")
	l.Fatal("msg")
}

func TestStringField(t *testing.T) {
	f := String("key", "value")
	assert.Equal(t, "key", f.GetKey())
	assert.Equal(t, "value", f.GetValue())
	assert.Equal(t, StringType, f.GetType())
}
