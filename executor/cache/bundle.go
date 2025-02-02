package cache

import (
	"context"
)

type sourceBundle struct {
	sources []Source
}

func (m *sourceBundle) Clean(ctx context.Context) {
	for _, source := range m.sources {
		source.Clean(ctx)
	}
}

func (m *sourceBundle) Get(ctx context.Context, statement string, statementArgs ...any) (any, error) {
	for _, source := range m.sources {
		val, err := source.Get(ctx, statement, statementArgs...)
		if err == nil {
			return val, nil
		}
	}
	return nil, ErrNotFound
}

func (m *sourceBundle) Set(ctx context.Context, cache any, statement string, statementArgs ...any) {
	for _, source := range m.sources {
		source.Set(ctx, cache, statement, statementArgs...)
	}
}

func NewModelBundle(opts ...Option) Source {
	b := &sourceBundle{}
	for _, opt := range opts {
		opt.apply(b)
	}
	return b
}
