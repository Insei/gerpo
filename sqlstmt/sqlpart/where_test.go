package sqlpart

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insei/fmap/v3"
	"github.com/stretchr/testify/assert"

	"github.com/insei/gerpo/types"
)

func (m *MockColumn) GetFilterFn(operation types.Operation) (func(ctx context.Context, value any) (string, bool, error), bool) {
	if !m.allowedAction {
		return nil, false
	}
	filters := map[types.Operation]func(ctx context.Context, value any) (string, bool, error){
		types.OperationEQ: func(ctx context.Context, value any) (string, bool, error) {
			if value == nil {
				return m.name + " IS NULL", false, nil
			}
			return m.name + " = ?", true, nil
		},
		types.OperationNEQ: func(ctx context.Context, value any) (string, bool, error) {
			if value == nil {
				return m.name + " IS NOT NULL", false, nil
			}
			return m.name + " != ?", true, nil
		},
		types.OperationGT: func(ctx context.Context, value any) (string, bool, error) {
			return m.name + " > ?", true, nil
		},
	}
	fn, ok := filters[operation]
	return fn, ok
}

func (m *MockColumn) GetField() fmap.Field {
	type MockStruct struct {
		Field string
	}
	stor, _ := fmap.Get[MockStruct]()
	return stor.MustFind("Field")
}

func TestWhereBuilder_StartAndEndGroup(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(builder *WhereBuilder)
		expectedSQL    string
		expectedValues []any
	}{
		{
			name: "Single condition within group",
			setup: func(builder *WhereBuilder) {
				builder.StartGroup()
				builder.AppendSQLWithValues("name = ?", true, "John")
				builder.EndGroup()
			},
			expectedSQL:    "(name = ?)",
			expectedValues: []any{"John"},
		},
		{
			name: "Multiple conditions within group",
			setup: func(builder *WhereBuilder) {
				builder.StartGroup()
				builder.AppendSQLWithValues("name = ?", true, "Alice")
				builder.AND()
				builder.AppendSQLWithValues("age > ?", true, 25)
				builder.EndGroup()
			},
			expectedSQL:    "(name = ? AND age > ?)",
			expectedValues: []any{"Alice", 25},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			builder := NewWhereBuilder(ctx)

			tc.setup(builder)

			if builder.sql != tc.expectedSQL {
				t.Errorf("Expected SQL '%s', got '%s'", tc.expectedSQL, builder.sql)
			}

			if !compareSlices(builder.Values(), tc.expectedValues) {
				t.Errorf("Expected values %v, got %v", tc.expectedValues, builder.Values())
			}
		})
	}
}

func TestWhereBuilder_AND_OR(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(builder *WhereBuilder)
		expectedSQL    string
		expectedValues []any
	}{
		{
			name: "Adding AND and OR conditions",
			setup: func(builder *WhereBuilder) {
				builder.AppendSQLWithValues("name = ?", true, "John")
				builder.AND()
				builder.AppendSQLWithValues("age > ?", true, 30)
				builder.OR()
				builder.AppendSQLWithValues("city = ?", true, "New York")
			},
			expectedSQL:    "name = ? AND age > ? OR city = ?",
			expectedValues: []any{"John", 30, "New York"},
		},
		{
			name: "Only AND conditions",
			setup: func(builder *WhereBuilder) {
				builder.AppendSQLWithValues("status = ?", true, "active")
				builder.AND()
				builder.AppendSQLWithValues("role = ?", true, "admin")
			},
			expectedSQL:    "status = ? AND role = ?",
			expectedValues: []any{"active", "admin"},
		},
		{
			name: "Only OR conditions",
			setup: func(builder *WhereBuilder) {
				builder.AppendSQLWithValues("country = ?", true, "USA")
				builder.OR()
				builder.AppendSQLWithValues("country = ?", true, "Canada")
			},
			expectedSQL:    "country = ? OR country = ?",
			expectedValues: []any{"USA", "Canada"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			builder := NewWhereBuilder(ctx)

			tc.setup(builder)

			if builder.sql != tc.expectedSQL {
				t.Errorf("Expected SQL '%s', got '%s'", tc.expectedSQL, builder.sql)
			}

			if !compareSlices(builder.Values(), tc.expectedValues) {
				t.Errorf("Expected values %v, got %v", tc.expectedValues, builder.Values())
			}
		})
	}
}

func TestWhereBuilder_AppendCondition(t *testing.T) {
	testCases := []struct {
		name           string
		column         types.Column
		operation      types.Operation
		value          any
		initialSQL     string
		expectedSQL    string
		expectedValues []any
		expectError    bool
	}{
		{
			name:           "Add EQ condition",
			column:         &MockColumn{name: "name", allowedAction: true},
			operation:      types.OperationEQ,
			value:          "Alice",
			initialSQL:     "",
			expectedSQL:    "name = ?",
			expectedValues: []any{"Alice"},
			expectError:    false,
		},
		{
			name:           "Add NEQ condition with nil",
			column:         &MockColumn{name: "deleted_at", allowedAction: true},
			operation:      types.OperationNEQ,
			value:          nil,
			initialSQL:     "",
			expectedSQL:    "deleted_at IS NOT NULL",
			expectedValues: []any{},
			expectError:    false,
		},
		{
			name:           "Attempt to add condition without allowed action",
			column:         &MockColumn{name: "password", allowedAction: false},
			operation:      types.OperationEQ,
			value:          "secret",
			initialSQL:     "",
			expectedSQL:    "",
			expectedValues: []any{},
			expectError:    true,
		},
		{
			name:           "Add multiple conditions",
			column:         &MockColumn{name: "age", allowedAction: true},
			operation:      types.OperationGT,
			value:          25,
			initialSQL:     "name = ?",
			expectedSQL:    "name = ? AND age > ?",
			expectedValues: []any{25},
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			builder := NewWhereBuilder(ctx)
			builder.sql = tc.initialSQL

			err := builder.AppendCondition(tc.column, tc.operation, tc.value)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if builder.sql != tc.expectedSQL {
				t.Errorf("Expected SQL '%s', got '%s'", tc.expectedSQL, builder.sql)
			}

			if !compareSlices(builder.Values(), tc.expectedValues) {
				t.Errorf("Expected values %v, got %v", tc.expectedValues, builder.Values())
			}
		})
	}
}

func TestWhereBuilder_SQL(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func(builder *WhereBuilder)
		expectedSQL string
	}{
		{
			name: "Empty SQL query",
			setup: func(builder *WhereBuilder) {
				// No operations
			},
			expectedSQL: "",
		},
		{
			name: "Single condition",
			setup: func(builder *WhereBuilder) {
				builder.AppendSQLWithValues("name = ?", true, "Bob")
			},
			expectedSQL: " WHERE name = ?",
		},
		{
			name: "Multiple conditions with AND and OR",
			setup: func(builder *WhereBuilder) {
				builder.AppendSQLWithValues("name = ?", true, "Bob")
				builder.AND()
				builder.AppendSQLWithValues("age >= ?", true, 20)
				builder.OR()
				builder.AppendSQLWithValues("city = ?", true, "Paris")
			},
			expectedSQL: " WHERE name = ? AND age >= ? OR city = ?",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			builder := NewWhereBuilder(ctx)
			tc.setup(builder)

			sql := builder.SQL()
			if sql != tc.expectedSQL {
				t.Errorf("Expected SQL '%s', got '%s'", tc.expectedSQL, sql)
			}
		})
	}
}

func TestWhereBuilder_Values(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(builder *WhereBuilder)
		expectedValues []any
	}{
		{
			name: "Single parameter",
			setup: func(builder *WhereBuilder) {
				builder.AppendSQLWithValues("name = ?", true, "Charlie")
			},
			expectedValues: []any{"Charlie"},
		},
		{
			name: "Multiple parameters with AND",
			setup: func(builder *WhereBuilder) {
				builder.AppendSQLWithValues("name = ?", true, "Charlie")
				builder.AND()
				builder.AppendSQLWithValues("age < ?", true, 40)
			},
			expectedValues: []any{"Charlie", 40},
		},
		{
			name: "Parameters with OR",
			setup: func(builder *WhereBuilder) {
				builder.AppendSQLWithValues("country = ?", true, "USA")
				builder.OR()
				builder.AppendSQLWithValues("country = ?", true, "Canada")
			},
			expectedValues: []any{"USA", "Canada"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			builder := NewWhereBuilder(ctx)
			tc.setup(builder)

			expected := tc.expectedValues
			actual := builder.Values()
			if !compareSlices(actual, expected) {
				t.Errorf("Expected values %v, got %v", expected, actual)
			}
		})
	}
}

// Helper function to compare two slices
func compareSlices(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
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
		name                     string
		query                    string
		value                    any
		expectedSQL              string
		expectedNeedAppendValues bool
	}{
		{
			name:                     "With valid slice",
			query:                    "fieldType",
			value:                    []any{1, 2, 3},
			expectedSQL:              "fieldType IN (?,?,?)",
			expectedNeedAppendValues: true,
		},
		{
			name:                     "With empty slice",
			query:                    "fieldType",
			value:                    []any{},
			expectedSQL:              "1 = 2",
			expectedNeedAppendValues: false,
		},
		{
			name:                     "With nil slice",
			query:                    "fieldType",
			value:                    ([]any)(nil),
			expectedSQL:              "1 = 2",
			expectedNeedAppendValues: false,
		},
		{
			name:                     "With nil",
			query:                    "fieldType",
			value:                    nil,
			expectedSQL:              "1 = 2",
			expectedNeedAppendValues: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genINFn(tc.query)
			sql, ok := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedNeedAppendValues, ok)
		})
	}
}

func TestGenNINFn(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name                     string
		query                    string
		value                    any
		expectedSQL              string
		expectedNeedAppendValues bool
	}{
		{
			name:                     "With valid slice",
			query:                    "fieldType",
			value:                    []any{1, 2, 3},
			expectedSQL:              "fieldType NOT IN (?,?,?)",
			expectedNeedAppendValues: true,
		},
		{
			name:                     "With empty slice",
			query:                    "fieldType",
			value:                    []any{},
			expectedSQL:              "1 = 1",
			expectedNeedAppendValues: false,
		},
		{
			name:                     "With nil slice",
			query:                    "fieldType",
			value:                    ([]any)(nil),
			expectedSQL:              "1 = 1",
			expectedNeedAppendValues: false,
		},
		{
			name:                     "With nil",
			query:                    "fieldType",
			value:                    nil,
			expectedSQL:              "1 = 1",
			expectedNeedAppendValues: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genNINFn(tc.query)
			sql, needAppendValues := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedNeedAppendValues, needAppendValues)
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
			expectedSQL: "LOWER(fieldType) LIKE LOWER(CONCAT('%', ?, '%'))",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := genCTFn(tc.query)
			sql, needAppendValues := fn(ctx, tc.value)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedOK, needAppendValues)
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
			expectedSQL: "LOWER(fieldType) NOT LIKE LOWER(CONCAT('%', ?, '%'))",
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
			expectedSQL: "LOWER(fieldType) LIKE LOWER(CONCAT(?, '%'))",
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
			expectedSQL: "LOWER(fieldType) NOT LIKE LOWER(CONCAT(?, '%'))",
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
			expectedSQL: "LOWER(fieldType) LIKE LOWER(CONCAT('%', ?))",
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
			expectedSQL: "LOWER(fieldType) NOT LIKE LOWER(CONCAT('%', ?))",
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
			type TestModel struct {
				Int     int
				Float64 float64
				String  string
				Bool    bool
				Time    time.Time
				UUID    uuid.UUID
				TimePtr *time.Time
			}
			fields, _ := fmap.Get[TestModel]()
			field := fields.MustFind(tc.fieldName)

			filters := GetFieldTypeFilters(field, "query")
			assert.Len(t, filters, len(tc.expectedOps))

			for _, op := range tc.expectedOps {
				assert.Contains(t, filters, op, "Filter for operation %v is missing", op)
			}
		})
	}
}
