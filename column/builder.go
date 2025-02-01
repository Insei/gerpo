package column

import (
	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/types"
)

type Builder struct {
	opts  []Option
	field fmap.Field
}

// NewBuilder creates and returns a new Builder instance initialized with the given fmap.Field.
func NewBuilder(field fmap.Field) *Builder {
	return &Builder{
		field: field,
	}
}

// WithAlias sets sql column alias.
func (b *Builder) WithAlias(alias string) *Builder {
	opt := WithAlias(alias)
	b.opts = append(b.opts, opt)
	return b
}

// WithTable sets the table name for the column.
func (b *Builder) WithTable(table string) *Builder {
	opt := WithTable(table)
	b.opts = append(b.opts, opt)
	return b
}

// WithColumnName sets the name of the SQL column.
func (b *Builder) WithColumnName(column string) *Builder {
	opt := WithColumnName(column)
	b.opts = append(b.opts, opt)
	return b
}

// WithInsertProtection appends an option to prevent the column from being included in SQL INSERT actions.
func (b *Builder) WithInsertProtection() *Builder {
	b.opts = append(b.opts, WithInsertProtection())
	return b
}

// WithUpdateProtection appends an option to prevent the column from being included in SQL UPDATE actions.
func (b *Builder) WithUpdateProtection() *Builder {
	b.opts = append(b.opts, WithUpdateProtection())
	return b
}

// Build constructs and returns a types.Column instance based on the field and options configured in the Builder.
func (b *Builder) Build() types.Column {
	return New(b.field, b.opts...)
}
