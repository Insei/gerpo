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
