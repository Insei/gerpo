package sqlpart

import (
	"context"
	"strings"
)

type Join interface {
	JOIN(joinFn func(ctx context.Context) string)
}

type JoinBuilder struct {
	ctx   context.Context
	joins []func(ctx context.Context) string
}

func NewJoinBuilder(ctx context.Context) *JoinBuilder { return &JoinBuilder{ctx: ctx} }

// Reset prepares the builder for reuse by a new query without dropping the underlying slice.
func (b *JoinBuilder) Reset(ctx context.Context) {
	b.ctx = ctx
	for i := range b.joins {
		b.joins[i] = nil
	}
	b.joins = b.joins[:0]
}

func (b *JoinBuilder) JOIN(joinFn func(ctx context.Context) string) {
	b.joins = append(b.joins, joinFn)
}

func (b *JoinBuilder) SQL() string {
	var sb strings.Builder
	for _, j := range b.joins {
		sb.WriteString(" " + j(b.ctx))
	}
	return sb.String()
}
