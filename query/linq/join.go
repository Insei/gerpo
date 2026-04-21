package linq

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
)

// JoinApplier lets JoinBuilder emit registered JOINs against a concrete
// sqlpart.Join buffer while having access to the per-request ctx — needed to
// invoke user-supplied resolvers that materialize the ON-clause arguments at
// the moment the statement runs.
type JoinApplier interface {
	Join() sqlpart.Join
	Ctx() context.Context
}

// JoinArgsResolver is invoked once per request. It receives the request-scoped
// ctx and must return the values that replace the `?` placeholders of the ON
// clause, in declaration order. Returning a non-nil error aborts the query —
// the error is propagated up through Apply without hitting the database.
type JoinArgsResolver = func(ctx context.Context) ([]any, error)

func NewJoinBuilder() *JoinBuilder {
	return &JoinBuilder{}
}

type joinKind uint8

const (
	joinLeftOn joinKind = iota
	joinInnerOn
)

func (k joinKind) String() string {
	switch k {
	case joinLeftOn:
		return "LEFT JOIN"
	case joinInnerOn:
		return "INNER JOIN"
	}
	return "JOIN"
}

type joinEntry struct {
	kind     joinKind
	table    string
	on       string
	resolver JoinArgsResolver // nil means a static JOIN with no bound arguments
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
		var args []any
		if e.resolver != nil {
			var err error
			args, err = e.resolver(applier.Ctx())
			if err != nil {
				return fmt.Errorf("join resolver (%s %s): %w", e.kind, e.table, err)
			}
		}
		switch e.kind {
		case joinLeftOn:
			j.JOINOn("LEFT JOIN "+e.table+" ON "+e.on, args...)
		case joinInnerOn:
			j.JOINOn("INNER JOIN "+e.table+" ON "+e.on, args...)
		}
	}
	return nil
}

// LeftJoinOn registers a LEFT JOIN with a fixed text. table is the joined
// table reference (`posts`, `posts AS p`, …); on is the raw ON-clause body
// where `?` placeholders refer to values returned by the optional resolver.
// Omit the resolver for a static JOIN without bound arguments. Passing more
// than one resolver panics — the API permits at most one.
func (q *JoinBuilder) LeftJoinOn(table, on string, resolver ...JoinArgsResolver) {
	q.entries = append(q.entries, joinEntry{
		kind:     joinLeftOn,
		table:    table,
		on:       on,
		resolver: pickResolver("LeftJoinOn", resolver),
	})
}

// InnerJoinOn is the INNER counterpart of LeftJoinOn with identical resolver
// semantics.
func (q *JoinBuilder) InnerJoinOn(table, on string, resolver ...JoinArgsResolver) {
	q.entries = append(q.entries, joinEntry{
		kind:     joinInnerOn,
		table:    table,
		on:       on,
		resolver: pickResolver("InnerJoinOn", resolver),
	})
}

func pickResolver(method string, resolver []JoinArgsResolver) JoinArgsResolver {
	switch len(resolver) {
	case 0:
		return nil
	case 1:
		return resolver[0]
	default:
		panic(fmt.Sprintf("gerpo: %s accepts at most one resolver, got %d", method, len(resolver)))
	}
}
