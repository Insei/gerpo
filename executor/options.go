package executor

import (
	"context"

	"github.com/insei/gerpo/executor/cache"
)

// SpanEnd is returned by a Tracer to signal the end of a span. The wrapped
// operation reports its terminal error (or nil on success) through it.
type SpanEnd func(err error)

// Tracer is gerpo's OpenTelemetry-friendly tracing hook. It is invoked at the
// start of every public Executor operation (GetOne, GetMultiple, Count,
// InsertOne, Update, Delete). The returned context is propagated downstream
// (so child operations land inside the span) and the SpanEnd is called with
// the operation's terminal error.
//
// gerpo deliberately does not import any tracing library — implement the
// adapter against your tracer of choice (OpenTelemetry, Datadog, OpenCensus, …).
type Tracer func(ctx context.Context, op string) (context.Context, SpanEnd)

type options struct {
	cacheSource cache.Storage
	tracer      Tracer
}

type Option interface {
	apply(c *options)
}

// optionFunc is a type that implements the Option interface.
type optionFn func(c *options)

// apply implements the Option interface for columnOptionFn.
// It calls the underlying function with the given *options.
func (f optionFn) apply(c *options) {
	f(c)
}

func WithCacheStorage(source cache.Storage) Option {
	return optionFn(func(o *options) {
		if source != nil {
			o.cacheSource = source
		}
	})
}

// WithTracer installs a tracing hook called around every Executor operation.
// Pass nil to disable tracing (this is the default). The hook signature is
// driver-agnostic — see Tracer for an adapter pattern compatible with
// OpenTelemetry / Datadog / OpenCensus.
func WithTracer(tracer Tracer) Option {
	return optionFn(func(o *options) {
		o.tracer = tracer
	})
}
