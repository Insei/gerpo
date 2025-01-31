package executor

import (
	"context"

	"github.com/insei/gerpo/cache"
)

func get[TCached any](ctx context.Context, b cache.Source, stmt string, stmtArgs ...any) (*TCached, bool) {
	if b == nil {
		return nil, false
	}
	cached, err := b.Get(ctx, stmt, stmtArgs...)
	if err == nil {
		cachedTyped, ok := cached.(TCached)
		// TODO: log failed cast
		if ok {
			return &cachedTyped, true
		}
	}
	return nil, false
}

func set(ctx context.Context, b cache.Source, cache any, statement string, statementArgs ...any) {
	if b == nil {
		return
	}
	b.Set(ctx, cache, statement, statementArgs...)
}

func clean(ctx context.Context, b cache.Source) {
	if b == nil {
		return
	}
	b.Clean(ctx)
}
