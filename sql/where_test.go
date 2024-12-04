package sql

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insei/fmap/v3"
	"github.com/stretchr/testify/assert"

	"github.com/insei/gerpo/types"
)

type TestModel struct {
	Int     int
	Float64 float64
	String  string
	Bool    bool
	Time    time.Time
	UUID    uuid.UUID
	TimePtr *time.Time
}

func TestGenEQFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With non-nil value",
			query:       "fieldType",
			value:       "value",
			expectedSQL: "fieldType = ?",
			expectedOK:  true,
		},
		{
			name:        "With nil value",
			query:       "fieldType",
			value:       nil,
			expectedSQL: "fieldType IS NULL",
			expectedOK:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genEQFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenNEQFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With non-nil value",
			query:       "fieldType",
			value:       "value",
			expectedSQL: "fieldType != ?",
			expectedOK:  true,
		},
		{
			name:        "With nil value",
			query:       "fieldType",
			value:       nil,
			expectedSQL: "fieldType IS NOT NULL",
			expectedOK:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genNEQFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenLTFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With value",
			query:       "fieldType",
			value:       10,
			expectedSQL: "fieldType < ?",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genLTFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenLTEFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With value",
			query:       "fieldType",
			value:       10,
			expectedSQL: "fieldType <= ?",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genLTEFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenGTFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With value",
			query:       "fieldType",
			value:       10,
			expectedSQL: "fieldType > ?",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genGTFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenGTEFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With value",
			query:       "fieldType",
			value:       10,
			expectedSQL: "fieldType >= ?",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genGTEFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenINFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With valid slice",
			query:       "fieldType",
			value:       []any{1, 2, 3},
			expectedSQL: "fieldType IN (?,?,?)",
			expectedOK:  true,
		},
		{
			name:        "With empty slice",
			query:       "fieldType",
			value:       []any{},
			expectedSQL: "",
			expectedOK:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genINFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenNINFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With valid slice",
			query:       "fieldType",
			value:       []any{1, 2, 3},
			expectedSQL: "fieldType NOT IN (?,?,?)",
			expectedOK:  true,
		},
		{
			name:        "With empty slice",
			query:       "fieldType",
			value:       []any{},
			expectedSQL: "",
			expectedOK:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genNINFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenCTFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With value",
			query:       "fieldType",
			value:       "value",
			expectedSQL: "LOWER(fieldType) LIKE LOWER('%' || ? || '%')",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genCTFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenNCTFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With value",
			query:       "fieldType",
			value:       "value",
			expectedSQL: "LOWER(fieldType) NOT LIKE LOWER('%' || ? || '%')",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genNCTFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenBWFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With value",
			query:       "fieldType",
			value:       "value",
			expectedSQL: "LOWER(fieldType) LIKE LOWER(? || '%')",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genBWFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenNBWFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With value",
			query:       "fieldType",
			value:       "value",
			expectedSQL: "LOWER(fieldType) NOT LIKE LOWER(? || '%')",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genNBWFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenEWFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With value",
			query:       "fieldType",
			value:       "value",
			expectedSQL: "LOWER(fieldType) LIKE LOWER('%' || ?)",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genEWFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGenNEWFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		value       any
		expectedSQL string
		expectedOK  bool
	}{
		{
			name:        "With value",
			query:       "fieldType",
			value:       "value",
			expectedSQL: "LOWER(fieldType) NOT LIKE LOWER('%' || ?)",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genNEWFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestGetDefaultTypeFilters(t *testing.T) {
	type testCase struct {
		name        string
		fieldName   string
		expectedOps []types.Operation
	}

	testCases := []testCase{
		{
			name:      "Boolean fieldType",
			fieldName: "Bool",
			expectedOps: []types.Operation{
				types.OperationEQ,
				types.OperationNEQ,
			},
		},
		{
			name:      "String fieldType",
			fieldName: "String",
			expectedOps: []types.Operation{
				types.OperationEQ,
				types.OperationNEQ,
				types.OperationIN,
				types.OperationNIN,
				types.OperationCT,
				types.OperationNCT,
				types.OperationBW,
				types.OperationNBW,
				types.OperationEW,
				types.OperationNEW,
			},
		},
		{
			name:      "Integer fieldType",
			fieldName: "Int",
			expectedOps: []types.Operation{
				types.OperationEQ,
				types.OperationNEQ,
				types.OperationLT,
				types.OperationLTE,
				types.OperationGT,
				types.OperationGTE,
				types.OperationIN,
				types.OperationNIN,
			},
		},
		{
			name:      "Float fieldType",
			fieldName: "Float64",
			expectedOps: []types.Operation{
				types.OperationEQ,
				types.OperationNEQ,
				types.OperationLT,
				types.OperationLTE,
				types.OperationGT,
				types.OperationGTE,
				types.OperationIN,
				types.OperationNIN,
			},
		},
		{
			name:      "Time fieldType",
			fieldName: "Time",
			expectedOps: []types.Operation{
				types.OperationLT,
				types.OperationGT,
			},
		},
		{
			name:      "UUID fieldType",
			fieldName: "UUID",
			expectedOps: []types.Operation{
				types.OperationEQ,
				types.OperationNEQ,
				types.OperationIN,
				types.OperationNIN,
			},
		},
		{
			name:      "Time ptr fieldType",
			fieldName: "TimePtr",
			expectedOps: []types.Operation{
				types.OperationEQ,
				types.OperationNEQ,
				types.OperationLT,
				types.OperationGT,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fields, _ := fmap.Get[TestModel]()
			field := fields.MustFind(tc.fieldName)

			filters := GetDefaultTypeFilters(field, "query")
			assert.Len(t, filters, len(tc.expectedOps))

			for _, op := range tc.expectedOps {
				assert.Contains(t, filters, op, "Filter for operation %v is missing", op)
			}
		})
	}
}

func TestSQLAndValues(t *testing.T) {
	builder := &StringWhereBuilder{
		sql:    "SELECT * FROM table WHERE",
		values: []any{"initial"},
	}

	assert.Equal(t, "SELECT * FROM table WHERE", builder.SQL())
	assert.Equal(t, []any{"initial"}, builder.Values())
}

func TestStartGroup(t *testing.T) {
	builder := &StringWhereBuilder{}

	builder.StartGroup()
	assert.Equal(t, "(", builder.SQL())
}

func TestEndGroup(t *testing.T) {
	builder := &StringWhereBuilder{}

	builder.EndGroup()
	assert.Equal(t, ")", builder.SQL())
}

func TestStartAndEndGroup(t *testing.T) {
	builder := &StringWhereBuilder{}

	builder.StartGroup()
	builder.EndGroup()
	assert.Equal(t, "()", builder.SQL())
}

func TestANDOperator(t *testing.T) {
	builder := &StringWhereBuilder{}

	builder.sql = "condition1"
	builder.AND()
	assert.Equal(t, "condition1 AND ", builder.SQL())
}

func TestOROperator(t *testing.T) {
	builder := &StringWhereBuilder{}

	builder.sql = "condition1"
	builder.OR()
	assert.Equal(t, "condition1 OR ", builder.SQL())
}

func TestAppendSQLWithValues(t *testing.T) {
	testCases := []struct {
		name           string
		initialSQL     string
		initialValues  []any
		sql            string
		appendValue    bool
		value          any
		expectedSQL    string
		expectedValues []any
	}{
		{
			name:           "Append SQL without value",
			initialSQL:     "SELECT * FROM table WHERE ",
			initialValues:  []any{},
			sql:            "field = ?",
			appendValue:    false,
			value:          nil,
			expectedSQL:    "SELECT * FROM table WHERE field = ?",
			expectedValues: []any{},
		},
		{
			name:           "Append SQL with value",
			initialSQL:     "SELECT * FROM table WHERE ",
			initialValues:  []any{},
			sql:            "field = ?",
			appendValue:    true,
			value:          "value",
			expectedSQL:    "SELECT * FROM table WHERE field = ?",
			expectedValues: []any{"value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := &StringWhereBuilder{
				sql:    tc.initialSQL,
				values: tc.initialValues,
			}

			builder.AppendSQLWithValues(tc.sql, tc.appendValue, tc.value)
			assert.Equal(t, tc.expectedSQL, builder.sql)
			assert.Equal(t, tc.expectedValues, builder.values)
		})
	}
}
