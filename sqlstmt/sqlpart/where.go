package sqlpart

import (
	"context"
	"fmt"
	"reflect"

	"github.com/insei/gerpo/types"
)

type Where interface {
	StartGroup()
	EndGroup()
	AND()
	OR()
	AppendSQLWithValues(sql string, appendValue bool, value any)
	AppendCondition(cl types.Column, operation types.Operation, val any) error
}

type WhereBuilder struct {
	ctx    context.Context
	sql    []byte
	values []any
}

func NewWhereBuilder(ctx context.Context) *WhereBuilder {
	return &WhereBuilder{
		ctx: ctx,
	}
}

// Reset prepares the builder for reuse by a new query without dropping underlying buffers.
func (b *WhereBuilder) Reset(ctx context.Context) {
	b.ctx = ctx
	b.sql = b.sql[:0]
	b.values = b.values[:0]
}

func (b *WhereBuilder) SQL() string {
	if len(b.sql) < 1 {
		return ""
	}
	return " WHERE " + string(b.sql)
}

func (b *WhereBuilder) Values() []any {
	return b.values
}

func (b *WhereBuilder) StartGroup() {
	if b.needANDBeforeCondition() {
		b.AND()
	}
	b.sql = append(b.sql, '(')
}
func (b *WhereBuilder) EndGroup() {
	b.sql = append(b.sql, ')')
}

func (b *WhereBuilder) AND() {
	b.sql = append(b.sql, " AND "...)
}

func (b *WhereBuilder) OR() {
	b.sql = append(b.sql, " OR "...)
}

func (b *WhereBuilder) appendValue(val any) {
	switch values := val.(type) {
	case []any:
		if len(values) > 0 {
			firstValTypeOf := reflect.ValueOf(values[0])
			if firstValTypeOf.Kind() == reflect.Slice {
				for i := 0; i < firstValTypeOf.Len(); i++ {
					b.values = append(b.values, firstValTypeOf.Index(i).Interface())
				}
			} else {
				b.values = append(b.values, values...)
			}
		}
	default:
		b.values = append(b.values, val)
	}
}

func (b *WhereBuilder) AppendSQLWithValues(sql string, appendValue bool, value any) {
	b.sql = append(b.sql, sql...)
	if appendValue {
		b.appendValue(value)
	}
}

func (b *WhereBuilder) needANDBeforeCondition() bool {
	if len(b.sql) < 4 {
		return false
	}
	endsWith := string(b.sql[len(b.sql)-4:])
	switch endsWith {
	case "AND ", " OR ":
		return false
	default:
		// group starts 'AND' no needed
		if endsWith[len(endsWith)-1:] == "(" {
			return false
		}
	}
	return true
}

// columnSQLArgsProvider mirrors sqlstmt.columnArgsProvider so the where-builder
// can prepend Compute-bound args without taking a dependency on the higher-level
// sqlstmt package.
type columnSQLArgsProvider interface {
	SQLArgs() []any
}

func (b *WhereBuilder) AppendCondition(cl types.Column, operation types.Operation, val any) error {
	if cl.IsAggregate() && !cl.HasFilterOverride(operation) {
		return fmt.Errorf("aggregate virtual column %q cannot be filtered without an explicit Filter() override (op=%s)",
			cl.GetField().GetStructPath(), operation)
	}
	filterFn, ok := cl.GetFilterFn(operation)
	if !ok {
		return fmt.Errorf("for field %s whereSQL %s option is not available", cl.GetField().GetStructPath(), operation)
	}
	sql, args, err := filterFn(b.ctx, val)
	if err != nil {
		return err
	}
	if sql == "" {
		return nil
	}
	if b.needANDBeforeCondition() {
		b.AND()
	}
	b.sql = append(b.sql, sql...)
	// Auto-derived filters wrap the column expression as `(compute_sql) op ?`, so any
	// bound args belonging to compute_sql must appear *before* the user value. Custom
	// filter overrides own their SQL entirely and decide whether to include those args.
	if !cl.HasFilterOverride(operation) {
		if ap, ok := cl.(columnSQLArgsProvider); ok {
			for _, a := range ap.SQLArgs() {
				b.appendValue(a)
			}
		}
	}
	for _, a := range args {
		b.appendValue(a)
	}
	return nil
}
