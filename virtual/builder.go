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

type BoolEQFilterBuilder struct {
	trueSQL, falseSQL, nilSQL func(ctx context.Context) string
}

func (b *BoolEQFilterBuilder) AddTrueSQLFn(fn func(ctx context.Context) string) *BoolEQFilterBuilder {
	b.trueSQL = fn
	return b
}
func (b *BoolEQFilterBuilder) AddFalseSQLFn(fn func(ctx context.Context) string) *BoolEQFilterBuilder {
	b.falseSQL = fn
	return b
}

func (b *BoolEQFilterBuilder) AddNilSQLFn(fn func(ctx context.Context) string) *BoolEQFilterBuilder {
	b.nilSQL = fn
	return b
}

func (b *BoolEQFilterBuilder) validate(field fmap.Field) {
	if field.GetDereferencedType().Kind() != reflect.Bool {
		panic(fmt.Errorf("bool filter is not applicable to %s field, types mismatch", field.GetStructPath()))
	}
	if field.GetType().Kind() == reflect.Ptr && nil == b.nilSQL {
		panic(fmt.Errorf("you need to add nilSQL to complete setup, because the %s field has reference boolean type", field.GetStructPath()))
	}
}

func (b *Builder) WithBoolEqFilter(fn func(b *BoolEQFilterBuilder)) *Builder {
	opt := WithBoolEqFilter(fn)
	b.opts = append(b.opts, opt)
	return b
}

func (b *Builder) Build() types.Column {
	return New(b.field, b.opts...)
}
