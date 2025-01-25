package sqlstmt

import (
	"context"
	"testing"

	"github.com/insei/gerpo/types"
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
		expectPanic    bool
	}{
		{
			name: "SQL Generation with One Column\n",
			setup: func(insert *Insert) {
				mockColumns := newMockExecutionColumns([]types.Column{
					&mockColumn{name: "id", hasName: true, allowedAction: true},
				})
				insert.columns = mockColumns
			},
			expectedSQL:    "INSERT INTO products (id) VALUES (?)",
			expectedValues: []any{},
			expectPanic:    false,
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
			expectPanic:    false,
		},
		{
			name: "SQL Generation with Missing Column Names",
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
			expectPanic:    false,
		},
		{
			name: "SQL Generation Without Columns",
			setup: func(insert *Insert) {
				mockColumns := newMockExecutionColumns([]types.Column{})
				insert.columns = mockColumns
			},
			expectedSQL:    "",
			expectedValues: []any{},
			expectPanic:    false,
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
			expectedSQL:    "INSERT INTO  (id) VALUES (?)",
			expectedValues: []any{},
			expectPanic:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tc.expectPanic {
						t.Errorf("Unexcpected panic: %v", r)
					}
				} else {
					if tc.expectPanic {
						t.Errorf("Expected panic, but not occurred")
					}
				}
			}()

			ctx := context.Background()
			insert := &Insert{
				ctx:   ctx,
				table: "products",
			}
			tc.setup(insert)
			insert.vals = newValues(insert.columns)

			sql, values := insert.SQL()

			if !tc.expectPanic {
				if sql != tc.expectedSQL {
					t.Errorf("Excpected SQL '%s', got '%s'", tc.expectedSQL, sql)
				}

				if !compareSlices(values, tc.expectedValues) {
					t.Errorf("Expected values %v, got %v", tc.expectedValues, values)
				}
			}
		})
	}
}
