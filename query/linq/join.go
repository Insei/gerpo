package linq

import (
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
	joinLeftOn joinKind = iota
	joinInnerOn
)

type joinEntry struct {
	kind  joinKind
	table string
	on    string
	args  []any
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
		case joinLeftOn:
			j.JOINOn("LEFT JOIN "+e.table+" ON "+e.on, e.args...)
		case joinInnerOn:
			j.JOINOn("INNER JOIN "+e.table+" ON "+e.on, e.args...)
		}
	}
	return nil
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

// InnerJoinOn is the parameter-bound counterpart of LeftJoinOn.
func (q *JoinBuilder) InnerJoinOn(table, on string, args ...any) {
	q.entries = append(q.entries, joinEntry{
		kind:  joinInnerOn,
		table: table,
		on:    on,
		args:  args,
	})
}
