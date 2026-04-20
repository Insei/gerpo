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

// OmitOnInsert excludes the column from every INSERT statement. Pair with a
// hook or a DB-side trigger if the value has to be set elsewhere. Typical use:
// UpdatedAt / DeletedAt columns managed by triggers or BeforeUpdate hooks.
func (b *Builder) OmitOnInsert() *Builder {
	b.opts = append(b.opts, WithOmitOnInsert())
	return b
}

// OmitOnUpdate excludes the column from every UPDATE SET clause. Typical use:
// CreatedAt, or a primary key that must never be moved after insert.
func (b *Builder) OmitOnUpdate() *Builder {
	b.opts = append(b.opts, WithOmitOnUpdate())
	return b
}

// ReadOnly makes the column invisible to both INSERT and UPDATE — SELECT-only.
// Equivalent to chaining OmitOnInsert and OmitOnUpdate. Use it for columns
// whose value is always produced by the database (PK with DEFAULT
// gen_random_uuid(), identity columns, virtual-style expressions declared
// outside the virtual column API).
func (b *Builder) ReadOnly() *Builder {
	b.opts = append(b.opts, WithOmitOnInsert(), WithOmitOnUpdate())
	return b
}

// Build constructs and returns a types.Column instance based on the field and options configured in the Builder.
func (b *Builder) Build() (types.Column, error) {
	return New(b.field, b.opts...)
}
