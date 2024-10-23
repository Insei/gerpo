package virtual

import (
	"context"
	"testing"

	"github.com/insei/fmap/v3"
	"github.com/stretchr/testify/assert"

	"github.com/insei/gerpo/types"
)

func TestWithSQL(t *testing.T) {
	expectedSQL := "SELECT * FROM table"

	opt := WithSQL(func(ctx context.Context) string {
		return expectedSQL
	})

	c := column{
		base: &types.ColumnBase{},
	}

	opt.apply(&c)

	assert.NotNil(t, c.base.ToSQL)
	assert.Equal(t, expectedSQL, c.base.ToSQL(context.Background()))
}

func TestWithBoolEqFilter(t *testing.T) {
	trueSQL := func(ctx context.Context) string { return "IS TRUE" }
	falseSQL := func(ctx context.Context) string { return "IS FALSE" }
	nilSQL := func(ctx context.Context) string { return "IS NULL" }

	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("BoolField")

	tests := []struct {
		name  string
		value any
		sql   string
	}{
		{
			name:  "Nil value",
			value: nil,
			sql:   "IS NULL",
		},
		{
			name:  "Pointer true",
			value: func() *bool { v := true; return &v }(),
			sql:   "IS TRUE",
		},
		{
			name:  "Pointer false",
			value: func() *bool { v := false; return &v }(),
			sql:   "IS FALSE",
		},
		{
			name:  "Value true",
			value: true,
			sql:   "IS TRUE",
		},
		{
			name:  "Value false",
			value: false,
			sql:   "IS FALSE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithBoolEqFilter(func(b *BoolEQFilterBuilder) {
				b.trueSQL = trueSQL
				b.falseSQL = falseSQL
				b.nilSQL = nilSQL
			})
			c := column{
				base: &types.ColumnBase{
					Field:   field,
					Filters: types.NewFilterManagerForField(field),
				},
			}
			opt.apply(&c)
			assert.NotNil(t, opt)
		})
	}
}
