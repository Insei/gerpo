package ctx

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	return storage.Get(s.key, buildKey(statement, statementArgs))
}

func (s *CtxCache) Set(ctx context.Context, cache any, statement string, statementArgs ...any) {
	storage, err := s.getStorage(ctx)
	if err != nil {
		return
	}
	storage.Set(s.key, buildKey(statement, statementArgs), cache)
}

// buildKey assembles a cache key from a SQL statement and its arguments without going
// through fmt.Sprintf. It mirrors the original "%s%v" encoding closely enough for
// equality comparison while avoiding the allocations fmt incurs for each argument.
func buildKey(statement string, args []any) string {
	var sb strings.Builder
	sb.Grow(len(statement) + 2 + len(args)*8)
	sb.WriteString(statement)
	sb.WriteByte('[')
	for i, a := range args {
		if i > 0 {
			sb.WriteByte(' ')
		}
		writeArg(&sb, a)
	}
	sb.WriteByte(']')
	return sb.String()
}

func writeArg(sb *strings.Builder, a any) {
	switch v := a.(type) {
	case nil:
		sb.WriteString("<nil>")
	case string:
		sb.WriteString(v)
	case int:
		sb.WriteString(strconv.Itoa(v))
	case int64:
		sb.WriteString(strconv.FormatInt(v, 10))
	case int32:
		sb.WriteString(strconv.FormatInt(int64(v), 10))
	case uint:
		sb.WriteString(strconv.FormatUint(uint64(v), 10))
	case uint64:
		sb.WriteString(strconv.FormatUint(v, 10))
	case uint32:
		sb.WriteString(strconv.FormatUint(uint64(v), 10))
	case bool:
		if v {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case []byte:
		sb.Write(v)
	case uuid.UUID:
		sb.WriteString(v.String())
	default:
		fmt.Fprint(sb, a)
	}
}

func (s *CtxCache) Clean(ctx context.Context) {
	storage, err := s.getStorage(ctx)
	if err != nil {
		return
	}
	storage.Clean(s.key)
}
