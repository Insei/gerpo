package cache

import (
	"context"
)

type modelBundle struct {
	sources []Source
}

func (m *modelBundle) Clean(ctx context.Context) {
	for _, source := range m.sources {
		source.Clean(ctx)
	}
}

func (m *modelBundle) Get(ctx context.Context, statement string, statementArgs ...any) (any, error) {
	for _, source := range m.sources {
		val, err := source.Get(ctx, statement, statementArgs...)
		if err == nil {
			return val, nil
		}
	}
	return nil, ErrNotFound
}

func (m *modelBundle) Set(ctx context.Context, cache any, statement string, statementArgs ...any) {
	for _, source := range m.sources {
		// TODO: Log error
		_ = source.Set(ctx, cache, statement, statementArgs...)
	}
}

func NewModelBundle(opts ...Option) ModelBundle {
	b := &modelBundle{}
	for _, opt := range opts {
		opt.apply(b)
	}
	return b
}
