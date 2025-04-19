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
	opts []func(JoinApplier) error
}

func (q *JoinBuilder) Apply(applier JoinApplier) error {
	for _, opt := range q.opts {
		err := opt(applier)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *JoinBuilder) LeftJoin(leftJoinFn func(ctx context.Context) string) {
	q.opts = append(q.opts, func(applier JoinApplier) error {
		applier.Join().JOIN(func(ctx context.Context) string {
			leftLoinStr := strings.TrimSpace(leftJoinFn(ctx))
			if leftLoinStr != "" {
				return "LEFT JOIN " + leftLoinStr
			}
			return ""
		})
		return nil
	})
}

func (q *JoinBuilder) InnerJoin(leftJoinFn func(ctx context.Context) string) {
	q.opts = append(q.opts, func(applier JoinApplier) error {
		applier.Join().JOIN(func(ctx context.Context) string {
			leftLoinStr := strings.TrimSpace(leftJoinFn(ctx))
			if leftLoinStr != "" {
				return "INNER JOIN " + leftLoinStr
			}
			return ""
		})
		return nil
	})
}
