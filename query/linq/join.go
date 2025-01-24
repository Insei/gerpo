package linq

import (
	"context"
	"strings"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
)

type JoinApplier interface {
	Join() sqlpart.Join
}

func NewJoinBuilder() *JoinBuilder {
	return &JoinBuilder{}
}

type JoinBuilder struct {
	opts []func(JoinApplier)
}

func (q *JoinBuilder) Apply(applier JoinApplier) {
	for _, opt := range q.opts {
		opt(applier)
	}
}

func (q *JoinBuilder) LeftJoin(leftJoinFn func(ctx context.Context) string) {
	q.opts = append(q.opts, func(applier JoinApplier) {
		applier.Join().JOIN(func(ctx context.Context) string {
			leftLoinStr := strings.TrimSpace(leftJoinFn(ctx))
			if leftLoinStr != "" {
				return "LEFT JOIN " + leftLoinStr
			}
			return ""
		})
	})
}
