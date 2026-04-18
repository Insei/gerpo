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

type joinKind uint8

const (
	joinLeft joinKind = iota
	joinInner
)

type joinEntry struct {
	kind joinKind
	fn   func(ctx context.Context) string
}

type JoinBuilder struct {
	entries []joinEntry
}

func (q *JoinBuilder) Apply(applier JoinApplier) error {
	if len(q.entries) == 0 {
		return nil
	}
	j := applier.Join()
	for i := range q.entries {
		e := &q.entries[i]
		kind := e.kind
		fn := e.fn
		j.JOIN(func(ctx context.Context) string {
			body := strings.TrimSpace(fn(ctx))
			if body == "" {
				return ""
			}
			switch kind {
			case joinLeft:
				return "LEFT JOIN " + body
			case joinInner:
				return "INNER JOIN " + body
			}
			return ""
		})
	}
	return nil
}

func (q *JoinBuilder) LeftJoin(leftJoinFn func(ctx context.Context) string) {
	q.entries = append(q.entries, joinEntry{kind: joinLeft, fn: leftJoinFn})
}

func (q *JoinBuilder) InnerJoin(innerJoinFn func(ctx context.Context) string) {
	q.entries = append(q.entries, joinEntry{kind: joinInner, fn: innerJoinFn})
}
