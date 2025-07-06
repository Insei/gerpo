package cache

import (
	"context"
)

type storagesBundle struct {
	storages []Storage
}

func (m *storagesBundle) Clean(ctx context.Context) {
	for _, source := range m.storages {
		source.Clean(ctx)
	}
}

func (m *storagesBundle) Get(ctx context.Context, statement string, statementArgs ...any) (any, error) {
	for _, storage := range m.storages {
		val, err := storage.Get(ctx, statement, statementArgs...)
		if err == nil {
			return val, nil
		}
	}
	return nil, ErrNotFound
}

func (m *storagesBundle) Set(ctx context.Context, cache any, statement string, statementArgs ...any) {
	for _, storage := range m.storages {
		storage.Set(ctx, cache, statement, statementArgs...)
	}
}

func NewModelBundle(opts ...Option) Storage {
	b := &storagesBundle{}
	for _, opt := range opts {
		opt.apply(b)
	}
	return b
}
