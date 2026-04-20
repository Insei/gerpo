package virtual

import (
	"context"
	"fmt"
	"testing"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func whereBuilderForTest() *sqlpart.WhereBuilder {
	return sqlpart.NewWhereBuilder(context.Background())
}

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

func TestBuilder_Compute_AutoDerivesFilters(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("NonBool") // string

	col, err := NewBuilder(field).Compute("LOWER(some_column)").Build()
	require.NoError(t, err)

	assert.Equal(t, "(LOWER(some_column))", col.ToSQL(context.Background()),
		"Compute wraps the expression in parentheses by contract")
	assert.False(t, col.IsAggregate())
	assert.False(t, col.HasFilterOverride(types.OperationEQ))

	fn, ok := col.GetFilterFn(types.OperationEQ)
	require.True(t, ok, "Compute auto-derives EQ for string fields")
	sql, args, err := fn(context.Background(), "abc")
	require.NoError(t, err)
	assert.Equal(t, "(LOWER(some_column)) = ?", sql)
	assert.Equal(t, []any{"abc"}, args)
}

func TestBuilder_Compute_StoresBoundArgs(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("NonBool")

	col, err := NewBuilder(field).Compute("EXTRACT(YEAR FROM created_at) = ?", 2026).Build()
	require.NoError(t, err)

	base := col.(*column).base
	assert.Equal(t, []any{2026}, base.SQLArgs,
		"Compute persists positional bound args on ColumnBase")
}

func TestBuilder_Aggregate_SkipsAutoFilters(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("NonBool")

	col, err := NewBuilder(field).Aggregate().Compute("SUM(x)").Build()
	require.NoError(t, err)

	assert.True(t, col.IsAggregate())
	_, ok := col.GetFilterFn(types.OperationEQ)
	assert.False(t, ok, "Aggregate columns must not auto-register filters")
}

func TestBuilder_Filter_OverridesSingleOperator(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("NonBool")

	col, err := NewBuilder(field).
		Compute("LOWER(name)").
		Filter(types.OperationEQ, Bound{SQL: "lower_name = ?"}).
		Build()
	require.NoError(t, err)

	assert.True(t, col.HasFilterOverride(types.OperationEQ))
	assert.False(t, col.HasFilterOverride(types.OperationLT))

	fn, ok := col.GetFilterFn(types.OperationEQ)
	require.True(t, ok)
	sql, args, err := fn(context.Background(), "abc")
	require.NoError(t, err)
	assert.Equal(t, "lower_name = ?", sql, "override replaces auto-derived EQ")
	assert.Equal(t, []any{"abc"}, args)

	fn, ok = col.GetFilterFn(types.OperationContainsFold)
	require.True(t, ok, "non-overridden operators stay auto-derived")
	sql, _, err = fn(context.Background(), "x")
	require.NoError(t, err)
	assert.Contains(t, sql, "(LOWER(name))")
}

func TestBuilder_Aggregate_GuardRejectsAutoFilter(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("NonBool")

	col, err := NewBuilder(field).Aggregate().Compute("SUM(amount)").Build()
	require.NoError(t, err)

	wb := whereBuilderForTest()
	err = wb.AppendCondition(col, types.OperationGT, 100)
	require.Error(t, err, "aggregate column without explicit Filter() must be rejected")
	assert.Contains(t, err.Error(), "aggregate")
}

func TestBuilder_Aggregate_AcceptsExplicitFilter(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("NonBool")

	col, err := NewBuilder(field).
		Aggregate().
		Compute("MAX(name)").
		Filter(types.OperationEQ, Bound{SQL: "MAX(name) = ?"}).
		Build()
	require.NoError(t, err)

	wb := whereBuilderForTest()
	require.NoError(t, wb.AppendCondition(col, types.OperationEQ, "alice"))
}

func TestBuilder_Filter_MatchSpec_BoolBranches(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("BoolField")

	col, err := NewBuilder(field).
		Aggregate(). // skip auto-filters so Filter is the only EQ implementation
		Compute("EXISTS (SELECT 1 FROM tokens WHERE user_id = users.id)").
		Filter(types.OperationEQ, Match{
			Cases: []MatchCase{
				{Value: true, Spec: SQL("EXISTS (SELECT 1 FROM tokens WHERE user_id = users.id)")},
				{Value: false, Spec: SQL("NOT EXISTS (SELECT 1 FROM tokens WHERE user_id = users.id)")},
			},
		}).
		Build()
	require.NoError(t, err)

	fn, ok := col.GetFilterFn(types.OperationEQ)
	require.True(t, ok)

	sqlTrue, _, err := fn(context.Background(), true)
	require.NoError(t, err)
	assert.Equal(t, "EXISTS (SELECT 1 FROM tokens WHERE user_id = users.id)", sqlTrue)

	sqlFalse, _, err := fn(context.Background(), false)
	require.NoError(t, err)
	assert.Equal(t, "NOT EXISTS (SELECT 1 FROM tokens WHERE user_id = users.id)", sqlFalse)
}
