package virtual

import (
	"context"
	"reflect"
	"strings"

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

func (b *Builder) WithSQL(fn func(ctx context.Context) string) *Builder {
	opt := WithSQL(fn)
	b.opts = append(b.opts, opt)
	return b
}

func (b *Builder) WithBoolEqFilter(trueSQL, falseSQL, nilSQL string) *Builder {
	if b.field.GetDereferencedType().Kind() != reflect.Bool {
		panic("Bool filter is not applicable to this field, types mismatch")
	}
	if b.field.GetType().Kind() == reflect.Ptr && "" == strings.TrimSpace(nilSQL) {
		panic("you need to add nilSQL to complete setup, because the field has reference boolean type")
	}
	opt := WithBoolEqFilter(trueSQL, falseSQL, nilSQL)
	b.opts = append(b.opts, opt)
	return b
}

func (b *Builder) Build() types.Column {
	return New(b.field, b.opts...)
}
