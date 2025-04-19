package sqlstmt

import (
	"context"
	"testing"

	"github.com/insei/gerpo/types"
	"github.com/stretchr/testify/assert"
)

func TestNewInsert(t *testing.T) {
	testCases := []struct {
		name     string
		table    string
		storage  types.ColumnsStorage
		expected string
	}{
		{
			name:     "Initialize Insert with valid table",
			table:    "products",
			storage:  newMockStorage(nil),
			expected: "products",
		},
		{
			name:     "Initialize Insert with empty string table",
			table:    "",
			storage:  newMockStorage(nil),
			expected: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			insert := NewInsert(ctx, tc.table, tc.storage)

			if insert.table != tc.expected {
				t.Errorf("Expected table '%s', got '%s'", tc.expected, insert.table)
			}

			if insert.ctx != ctx {
				t.Errorf("Expected context '%v', got '%v'", ctx, insert.ctx)
			}
		})
	}
}

func TestInsert_SQL(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(insert *Insert)
		expectedSQL    string
		expectedValues []any
		expectError    bool
	}{
		{
			name: "SQL Generation with One AsColumn\n",
			setup: func(insert *Insert) {
				mockColumns := newMockExecutionColumns([]types.Column{
					&mockColumn{name: "id", hasName: true, allowedAction: true},
				})
				insert.columns = mockColumns
			},
			expectedSQL:    "INSERT INTO products (id) VALUES (?)",
			expectedValues: []any{},
		},
		{
			name: "SQL Generation with Multiple Columns",
			setup: func(insert *Insert) {
				mockColumns := newMockExecutionColumns([]types.Column{
					&mockColumn{name: "id", hasName: true, allowedAction: true},
					&mockColumn{name: "name", hasName: true, allowedAction: true},
					&mockColumn{name: "price", hasName: true, allowedAction: true},
				})
				insert.columns = mockColumns
			},
			expectedSQL:    "INSERT INTO products (id, name, price) VALUES (?,?,?)",
			expectedValues: []any{},
		},
		{
			name: "SQL Generation with Missing AsColumn Names",
			setup: func(insert *Insert) {
				mockColumns := newMockExecutionColumns([]types.Column{
					&mockColumn{name: "id", hasName: true, allowedAction: true},
					&mockColumn{name: "", hasName: false, allowedAction: true},
					&mockColumn{name: "price", hasName: true, allowedAction: true},
				})
				insert.columns = mockColumns
			},
			expectedSQL:    "INSERT INTO products (id, price) VALUES (?,?)",
			expectedValues: []any{},
		},
		{
			name: "SQL Generation Without Columns",
			setup: func(insert *Insert) {
				mockColumns := newMockExecutionColumns([]types.Column{})
				insert.columns = mockColumns
			},
			expectError: true,
		},
		{
			name: "SQL Generation with an Empty Table",
			setup: func(insert *Insert) {
				insert.table = ""
				mockColumns := newMockExecutionColumns([]types.Column{
					&mockColumn{name: "id", hasName: true, allowedAction: true},
				})
				insert.columns = mockColumns
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			insert := &Insert{
				ctx:   ctx,
				table: "products",
			}
			tc.setup(insert)
			insert.vals = newValues(insert.columns)

			sql, values, err := insert.SQL()
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if sql != tc.expectedSQL {
				t.Errorf("Excpected SQL '%s', got '%s'", tc.expectedSQL, sql)
			}

			if !compareSlices(values, tc.expectedValues) {
				t.Errorf("Expected values %v, got %v", tc.expectedValues, values)
			}
		})
	}
}
