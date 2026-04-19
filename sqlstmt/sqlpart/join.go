package sqlpart

import (
	"context"
	"strings"
)

// Join is the interface gerpo's query layer talks to when adding JOIN clauses
// to a SELECT/UPDATE/DELETE statement.
type Join interface {
	// JOIN registers a callback whose returned text is inlined verbatim into
	// the SQL. The callback receives the request context. Values are NOT
	// parameterised — anything interpolated through the body lands in the SQL
	// as text. Prefer JOINOn when bound parameters are needed.
	JOIN(joinFn func(ctx context.Context) string)

	// JOINOn registers a JOIN whose text is fixed and whose parameters are
	// bound through the driver's placeholder mechanism, identical to WHERE
	// arguments. The provided sql must already start with the JOIN keyword
	// (e.g. "LEFT JOIN posts ON posts.user_id = users.id AND posts.tenant_id = ?").
	JOINOn(sql string, args ...any)
}

type joinPart struct {
	fn    func(ctx context.Context) string
	sql   string
	args  []any
	bound bool
}

type JoinBuilder struct {
	ctx    context.Context
	joins  []joinPart
	values []any
}

func NewJoinBuilder(ctx context.Context) *JoinBuilder { return &JoinBuilder{ctx: ctx} }

// Reset prepares the builder for reuse by a new query without dropping the underlying slice.
func (b *JoinBuilder) Reset(ctx context.Context) {
	b.ctx = ctx
	for i := range b.joins {
		b.joins[i] = joinPart{}
	}
	b.joins = b.joins[:0]
	for i := range b.values {
		b.values[i] = nil
	}
	b.values = b.values[:0]
}

func (b *JoinBuilder) JOIN(joinFn func(ctx context.Context) string) {
	b.joins = append(b.joins, joinPart{fn: joinFn})
}

func (b *JoinBuilder) JOINOn(sql string, args ...any) {
	b.joins = append(b.joins, joinPart{sql: sql, args: args, bound: true})
	if len(args) > 0 {
		b.values = append(b.values, args...)
	}
}

// Values returns the accumulated bound JOIN arguments in registration order.
// Callers must prepend these to WHERE values when building the final argument
// list, because JOIN clauses appear before WHERE in the generated SQL.
func (b *JoinBuilder) Values() []any {
	return b.values
}

func (b *JoinBuilder) SQL() string {
	var sb strings.Builder
	for i := range b.joins {
		j := &b.joins[i]
		var body string
		if j.bound {
			body = j.sql
		} else if j.fn != nil {
			body = j.fn(b.ctx)
		}
		if body == "" {
			continue
		}
		sb.WriteByte(' ')
		sb.WriteString(body)
	}
	return sb.String()
}
