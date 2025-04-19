package virtual

import (
	"context"
	"fmt"
	"testing"

	"github.com/insei/fmap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestModel struct {
	Active    *bool
	NonBool   string
	BoolField bool
}

func TestNewBuilder(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Active")

	t.Run("Test NewBuilder", func(t *testing.T) {
		builder := NewBuilder(field)
		assert.NotNil(t, builder)
		assert.Equal(t, field, builder.field)
	})
}

func TestBuilderWithSQL(t *testing.T) {
	trueSQL := func(ctx context.Context) string { return "IS TRUE" }

	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Active")

	t.Run("Test WithSQL", func(t *testing.T) {
		builder := &Builder{
			field: field,
		}
		builder.WithSQL(trueSQL)
		assert.Equal(t, 1, len(builder.opts))
	})
}

func TestBoolEQFilterBuilderAddTrueSQLFn(t *testing.T) {
	boolFilterBuilder := &BoolEQFilterBuilder{}

	trueSQL := func(ctx context.Context) string { return "IS TRUE" }

	t.Run("Test AddTrueSQLFn", func(t *testing.T) {
		result := boolFilterBuilder.AddTrueSQLFn(trueSQL)
		assert.NotNil(t, result)
	})
}

func TestBoolEQFilterBuilderAAddFalseSQLFn(t *testing.T) {
	boolFilterBuilder := &BoolEQFilterBuilder{}

	falseSQL := func(ctx context.Context) string { return "IS FALSE" }

	t.Run("Test AddFalseSQLFn", func(t *testing.T) {
		result := boolFilterBuilder.AddFalseSQLFn(falseSQL)
		assert.NotNil(t, result)
	})
}

func TestBoolEQFilterBuilderAddNilSQLFn(t *testing.T) {
	boolFilterBuilder := &BoolEQFilterBuilder{}

	nilSQL := func(ctx context.Context) string { return "IS NULL" }

	t.Run("Test AddNilSQLFn", func(t *testing.T) {
		result := boolFilterBuilder.AddNilSQLFn(nilSQL)
		assert.NotNil(t, result)
	})
}

func TestBoolEQFilterBuilderValidate(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	fieldBool := fields.MustFind("BoolField")
	fieldPtrBool := fields.MustFind("Active")
	fieldNonBool := fields.MustFind("NonBool")

	t.Run("should panic for non-bool field", func(t *testing.T) {
		boolFilterBuilder := &BoolEQFilterBuilder{}
		expectedErr := fmt.Errorf("bool query is not applicable to %s field, types mismatch", fieldNonBool.GetStructPath())
		assert.PanicsWithError(t, expectedErr.Error(), func() {
			boolFilterBuilder.validate(fieldNonBool)
		})
	})

	t.Run("should panic for pointer bool field without nilSQL", func(t *testing.T) {
		boolFilterBuilder := &BoolEQFilterBuilder{}
		expectedErr := fmt.Errorf("you need to add nilSQL to complete setup, because the %s field has reference boolean type", fieldPtrBool.GetStructPath())
		assert.PanicsWithError(t, expectedErr.Error(), func() {
			boolFilterBuilder.validate(fieldPtrBool)
		})
	})

	t.Run("should not panic for bool field", func(t *testing.T) {
		boolFilterBuilder := &BoolEQFilterBuilder{}
		require.NotPanics(t, func() {
			boolFilterBuilder.validate(fieldBool)
		})
	})

	t.Run("should not panic for pointer bool field with nilSQL", func(t *testing.T) {
		nilSQL := func(ctx context.Context) string { return "IS NULL" }
		boolFilterBuilder := &BoolEQFilterBuilder{
			nilSQL: nilSQL,
		}
		require.NotPanics(t, func() {
			boolFilterBuilder.validate(fieldPtrBool)
		})
	})
}

func TestBuilder_WithBoolEqFilter(t *testing.T) {
	builder := &Builder{}

	boolEqFn := func(b *BoolEQFilterBuilder) {
		b.AddTrueSQLFn(func(ctx context.Context) string { return "IS TRUE" })
	}

	t.Run("Test WithBoolEqFilter", func(t *testing.T) {
		result := builder.WithBoolEqFilter(boolEqFn)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(builder.opts))
	})
}

func TestBuilderBuild(t *testing.T) {

	t.Run("Test Build", func(t *testing.T) {
		fields, _ := fmap.Get[TestModel]()
		field := fields.MustFind("Active")

		builder := &Builder{
			field: field,
		}

		col, err := builder.Build()
		assert.NoError(t, err)
		assert.NotNil(t, col)
	})

	t.Run("Test Build nil field", func(t *testing.T) {
		builder := &Builder{
			field: nil,
		}
		col, err := builder.Build()
		assert.Error(t, err)
		assert.Nil(t, col)
	})
}
