package column

import (
	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/types"
)

type Builder struct {
	opts  []Option
	field fmap.Field
}

func NewBuilder(field fmap.Field) *Builder {
	return &Builder{
		field: field,
	}
}

func (b *Builder) WithAlias(alias string) *Builder {
	opt := WithAlias(alias)
	b.opts = append(b.opts, opt)
	return b
}

func (b *Builder) WithTable(table string) *Builder {
	opt := WithTable(table)
	b.opts = append(b.opts, opt)
	return b
}
func (b *Builder) WithColumnName(column string) *Builder {
	opt := WithColumnName(column)
	b.opts = append(b.opts, opt)
	return b
}

func (b *Builder) Build() types.Column {
	return New(b.field, b.opts...)
}
