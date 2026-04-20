package virtual

import (
	"context"
	"fmt"
	"reflect"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/types"
)

type Builder struct {
	opts  []Option
	field fmap.Field
}

// NewBuilder initializes and returns a new Builder instance for the specified field.
func NewBuilder(field fmap.Field) *Builder {
	return &Builder{
		field: field,
	}
}

// WithSQL adds a custom SQL generation function to the builder, modifying how the SQL statement is constructed.
//
// Deprecated: use Compute(sql, args...). The new method covers the same use case for static expressions
// and adds bound-args support; see docs/features/virtual-columns.md.
func (b *Builder) WithSQL(fn func(ctx context.Context) string) *Builder {
	opt := WithSQL(fn)
	b.opts = append(b.opts, opt)
	return b
}

// Compute sets a static SQL expression for the virtual column. The expression is always
// wrapped in parentheses automatically — that is part of the contract, not magic — so it
// composes cleanly inside larger predicates. Optional bound args travel with the column
// wherever it is referenced (SELECT/WHERE/ORDER).
//
// When the column is not Aggregate, standard operators (EQ, LT, IN, ...) are auto-derived
// from the field type, the same way they work for plain columns.
func (b *Builder) Compute(sql string, args ...any) *Builder {
	b.opts = append(b.opts, WithCompute(sql, args...))
	return b
}

// Aggregate marks the column as an aggregate expression (SUM, COUNT, ...). Aggregate
// columns reject WHERE filtering unless the operator has an explicit Filter override —
// the WhereBuilder returns an error to prevent silently invalid SQL.
func (b *Builder) Aggregate() *Builder {
	b.opts = append(b.opts, WithAggregate())
	return b
}

// Filter registers a custom filter for one operation. spec is a FilterSpec — see
// virtual.SQL / Bound / SQLArgs / Match / Func. Other operators keep their auto-derived
// implementations (unless the column is Aggregate).
func (b *Builder) Filter(op types.Operation, spec FilterSpec) *Builder {
	b.opts = append(b.opts, WithFilter(op, spec))
	return b
}

// BoolEQFilterBuilder is a builder for constructing boolean equality filters in SQL queries.
// It allows specifying SQL construction functions for true, false, and nil values.
type BoolEQFilterBuilder struct {
	trueSQL, falseSQL, nilSQL func(ctx context.Context) string
}

// AddTrueSQLFn sets the trueSQL function for the BoolEQFilterBuilder.
// Deprecated: this method can be changed soon or removed.
func (b *BoolEQFilterBuilder) AddTrueSQLFn(fn func(ctx context.Context) string) *BoolEQFilterBuilder {
	b.trueSQL = fn
	return b
}

// AddFalseSQLFn sets a custom SQL function for the false condition in a boolean filter.
// Deprecated: this method can be changed soon or removed.
func (b *BoolEQFilterBuilder) AddFalseSQLFn(fn func(ctx context.Context) string) *BoolEQFilterBuilder {
	b.falseSQL = fn
	return b
}

// AddNilSQLFn sets a function to generate SQL for nil boolean values.
// Deprecated: this method can be changed soon or removed.
func (b *BoolEQFilterBuilder) AddNilSQLFn(fn func(ctx context.Context) string) *BoolEQFilterBuilder {
	b.nilSQL = fn
	return b
}

// validate ensures the provided field is of boolean type and checks for nilSQL if the field is a pointer to bool.
func (b *BoolEQFilterBuilder) validate(field fmap.Field) {
	if field.GetDereferencedType().Kind() != reflect.Bool {
		panic(fmt.Errorf("bool query is not applicable to %s field, types mismatch", field.GetStructPath()))
	}
	if field.GetType().Kind() == reflect.Ptr && nil == b.nilSQL {
		panic(fmt.Errorf("you need to add nilSQL to complete setup, because the %s field has reference boolean type", field.GetStructPath()))
	}
}

// WithBoolEqFilter adds a boolean equality filter to the builder using a provided configuration function.
//
// Deprecated: use Filter(types.OperationEQ, virtual.Match{Cases: ..., Default: ...}) — the
// Match spec covers the true/false/nil branching declaratively without a callback.
func (b *Builder) WithBoolEqFilter(fn func(b *BoolEQFilterBuilder)) *Builder {
	opt := WithBoolEqFilter(fn)
	b.opts = append(b.opts, opt)
	return b
}

// Build constructs and returns an instance of types.Column based on the current field and options in the Builder.
func (b *Builder) Build() (types.Column, error) {
	return New(b.field, b.opts...)
}
