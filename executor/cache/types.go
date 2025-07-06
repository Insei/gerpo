package cache

import (
	"context"

	"github.com/insei/gerpo/executor/cache/types"
)

var (
	ErrNotFound           = types.ErrNotFound
	ErrWrongConfiguration = types.ErrWrongConfiguration
)

type Storage interface {
	Clean(ctx context.Context)
	Get(ctx context.Context, statement string, statementArgs ...any) (any, error)
	Set(ctx context.Context, cache any, statement string, statementArgs ...any)
}
