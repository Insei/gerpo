package ctx

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/insei/gerpo/executor/cache/types"
	"github.com/insei/gerpo/logger"
)

type CtxCache struct {
	key string
	log logger.Logger
}

func New(opts ...Option) *CtxCache {
	s := &CtxCache{
		log: logger.NoopLogger,
		key: uuid.New().String(),
	}
	for _, opt := range opts {
		opt.apply(s)
	}
	return s
}

func (s *CtxCache) getStorage(ctx context.Context) (*cacheStorage, error) {
	if ctx == nil {
		return nil, types.ErrNotFound
	}
	storage, ok := ctx.Value(ctxCacheKey).(*cacheStorage)
	if !ok || storage == nil {
		s.log.Ctx(ctx).Warn("not found",
			logger.String("driver", "ctx"),
			logger.String("details", "missing storage in context, miss middleware?"))
		return nil, types.ErrWrongConfiguration
	}
	return storage, nil
}

func (s *CtxCache) Get(ctx context.Context, statement string, statementArgs ...any) (any, error) {
	storage, err := s.getStorage(ctx)
	if err != nil {
		return nil, err
	}
	return storage.Get(s.key, fmt.Sprintf("%s%v", statement, statementArgs))
}

func (s *CtxCache) Set(ctx context.Context, cache any, statement string, statementArgs ...any) {
	storage, err := s.getStorage(ctx)
	if err != nil {
		return
	}
	storage.Set(s.key, fmt.Sprintf("%s%v", statement, statementArgs), cache)
}

func (s *CtxCache) Clean(ctx context.Context) {
	storage, err := s.getStorage(ctx)
	if err != nil {
		return
	}
	storage.Clean(s.key)
}
