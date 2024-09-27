package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/insei/fmap/v3"
)

func TestName(t *testing.T) {
	tt := &test{}
	fields, _ := fmap.GetFrom(tt)
	b := &columnBuilder[test]{
		builders: make([]ColumnBuilder, 0),
		fields:   fields,
		model:    tt,
	}
	b.
		Virtual(func(m *test) any {
			return &m.Bool
		}).
		WithSQL(func(ctx context.Context) string {
			return ""
		}).
		WithBoolEqFilter("", "", "")
	b.Column(func(m *test) any { return &m.Name })
	cls := b.Build()
	fmt.Println(cls)
}
