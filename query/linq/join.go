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
	joinLeftOn
	joinInnerOn
)

type joinEntry struct {
	kind  joinKind
	fn    func(ctx context.Context) string // legacy callback variants
	table string                           // *On variants
	on    string                           // *On variants
	args  []any                            // *On variants
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
		switch e.kind {
		case joinLeft:
			fn := e.fn
			j.JOIN(func(ctx context.Context) string {
				body := strings.TrimSpace(fn(ctx))
				if body == "" {
					return ""
				}
				return "LEFT JOIN " + body
			})
		case joinInner:
			fn := e.fn
			j.JOIN(func(ctx context.Context) string {
				body := strings.TrimSpace(fn(ctx))
				if body == "" {
					return ""
				}
				return "INNER JOIN " + body
			})
		case joinLeftOn:
			j.JOINOn("LEFT JOIN "+e.table+" ON "+e.on, e.args...)
		case joinInnerOn:
			j.JOINOn("INNER JOIN "+e.table+" ON "+e.on, e.args...)
		}
	}
	return nil
}

// LeftJoin registers a LEFT JOIN whose body is produced by a context-aware
// callback. Anything inside the returned string is inlined verbatim — values
// are NOT parameterised.
//
// Deprecated: prefer LeftJoinOn for safer, parameter-bound JOINs.
func (q *JoinBuilder) LeftJoin(leftJoinFn func(ctx context.Context) string) {
	q.entries = append(q.entries, joinEntry{kind: joinLeft, fn: leftJoinFn})
}

// InnerJoin registers an INNER JOIN whose body is produced by a context-aware
// callback. Same caveat as LeftJoin: no parameter binding.
//
// Deprecated: prefer InnerJoinOn for safer, parameter-bound JOINs.
func (q *JoinBuilder) InnerJoin(innerJoinFn func(ctx context.Context) string) {
	q.entries = append(q.entries, joinEntry{kind: joinInner, fn: innerJoinFn})
}

// LeftJoinOn registers a LEFT JOIN with a fixed text and bound arguments.
// table is the joined table reference (`posts`, `posts AS p`, …); on is the
// raw ON-clause body where `?` placeholders refer to args in declaration order.
// Bound arguments flow through the driver and avoid SQL injection.
func (q *JoinBuilder) LeftJoinOn(table, on string, args ...any) {
	q.entries = append(q.entries, joinEntry{
		kind:  joinLeftOn,
		table: table,
		on:    on,
		args:  args,
	})
}

// InnerJoinOn is the parameter-bound counterpart of InnerJoin.
func (q *JoinBuilder) InnerJoinOn(table, on string, args ...any) {
	q.entries = append(q.entries, joinEntry{
		kind:  joinInnerOn,
		table: table,
		on:    on,
		args:  args,
	})
}
