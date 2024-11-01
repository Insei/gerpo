package types

import "context"

type Source interface {
	Clean(ctx context.Context)
	Get(ctx context.Context, statement string, statementArgs ...any) (any, error)
	Set(ctx context.Context, cache any, statement string, statementArgs ...any) error
}
