package virtual

import (
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

// Build constructs and returns an instance of types.Column based on the current field and options in the Builder.
func (b *Builder) Build() (types.Column, error) {
	return New(b.field, b.opts...)
}
