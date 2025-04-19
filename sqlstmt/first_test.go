package sqlstmt

import (
	"context"
	"testing"

	"github.com/insei/gerpo/types"
	"github.com/stretchr/testify/assert"
)

func TestNewGetFirst(t *testing.T) {
	testCases := []struct {
		name     string
		table    string
		storage  types.ColumnsStorage
		expected string
	}{
		{
			name:     "Initialize GetFirst with valid table",
			table:    "products",
			storage:  newMockStorage(nil),
			expected: "products",
		},
		{
			name:     "Initialize GetFirst with empty table",
			table:    "",
			storage:  newMockStorage(nil),
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			getFirst := NewGetFirst(ctx, tc.table, tc.storage)

			if getFirst.table != tc.expected {
				t.Errorf("Expected table '%s', got '%s'", tc.expected, getFirst.table)
			}

			if getFirst.ctx != ctx {
				t.Errorf("Expected context '%v', got '%v'", ctx, getFirst.ctx)
			}

			if getFirst.sqlselect.columnsStorage != tc.storage {
				t.Errorf("Expected ColumnsStorage '%v', got '%v'", tc.storage, getFirst.sqlselect.columnsStorage)
			}
		})
	}
}

func TestGetFirst_SQL(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(getFirst *GetFirst)
		expectedSQL    string
		expectedValues []any
		expectError    bool
	}{
		{
			name: "Generate SQL with single column",
			setup: func(getFirst *GetFirst) {
				mockColumns := newMockExecutionColumns([]types.Column{
					&mockColumn{name: "id", allowedAction: true},
				})
				getFirst.columns = mockColumns
			},
			expectedSQL:    "SELECT id FROM products LIMIT 1",
			expectedValues: []any{},
		},
		{
			name: "Generate SQL with multiple columns",
			setup: func(getFirst *GetFirst) {
				mockColumns := newMockExecutionColumns([]types.Column{
					&mockColumn{name: "id", allowedAction: true},
					&mockColumn{name: "name", allowedAction: true},
					&mockColumn{name: "price", allowedAction: true},
				})
				getFirst.columns = mockColumns
			},
			expectedSQL:    "SELECT id, name, price FROM products LIMIT 1",
			expectedValues: []any{},
		},
		{
			name: "Generate SQL with WHERE condition",
			setup: func(getFirst *GetFirst) {
				mockColumns := newMockExecutionColumns([]types.Column{
					&mockColumn{name: "id", allowedAction: true},
					&mockColumn{name: "name", allowedAction: true},
				})
				getFirst.columns = mockColumns
				getFirst.where.AppendSQLWithValues("price > ?", true, 100)
			},
			expectedSQL:    "SELECT id, name FROM products WHERE price > ? LIMIT 1",
			expectedValues: []any{100},
		},
		{
			name: "Generate SQL with multiple WHERE conditions",
			setup: func(getFirst *GetFirst) {
				mockColumns := newMockExecutionColumns([]types.Column{
					&mockColumn{name: "id", allowedAction: true},
					&mockColumn{name: "name", allowedAction: true},
					&mockColumn{name: "price", allowedAction: true},
				})
				getFirst.columns = mockColumns
				getFirst.where.AppendSQLWithValues("price > ?", true, 50)
				getFirst.where.AND()
				getFirst.where.AppendSQLWithValues("stock > ?", true, 20)
			},
			expectedSQL:    "SELECT id, name, price FROM products WHERE price > ? AND stock > ? LIMIT 1",
			expectedValues: []any{50, 20},
		},
		{
			name: "Generate SQL with JOIN",
			setup: func(getFirst *GetFirst) {
				mockColumns := newMockExecutionColumns([]types.Column{
					&mockColumn{name: "products.id", allowedAction: true},
					&mockColumn{name: "categories.name", allowedAction: true},
				})
				getFirst.columns = mockColumns
				getFirst.join.JOIN(func(ctx context.Context) string {
					return "INNER JOIN categories ON products.category_id = categories.id"
				})
			},
			expectedSQL:    "SELECT products.id, categories.name FROM products INNER JOIN categories ON products.category_id = categories.id LIMIT 1",
			expectedValues: []any{},
		},
		{
			name: "Generate SQL with WHERE and JOIN",
			setup: func(getFirst *GetFirst) {
				mockColumns := newMockExecutionColumns([]types.Column{
					&mockColumn{name: "products.id", allowedAction: true},
					&mockColumn{name: "categories.name", allowedAction: true},
				})
				getFirst.columns = mockColumns
				getFirst.join.JOIN(func(ctx context.Context) string {
					return "LEFT JOIN categories ON products.category_id = categories.id"
				})
				getFirst.where.AppendSQLWithValues("categories.active = ?", true, true)
			},
			expectedSQL:    "SELECT products.id, categories.name FROM products LEFT JOIN categories ON products.category_id = categories.id WHERE categories.active = ? LIMIT 1",
			expectedValues: []any{true},
		},
		{
			name: "Panic when no columns are set",
			setup: func(getFirst *GetFirst) {
				// Не устанавливаем колонки, чтобы вызвать панику
			},
			expectedSQL:    "",
			expectedValues: []any{},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		tc := tc // локальная копия для использования в defer/recover
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			getFirst := NewGetFirst(ctx, "products", newMockStorage([]types.Column{}))

			tc.setup(getFirst)

			sql, values, err := getFirst.SQL()
			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if sql != tc.expectedSQL {
				t.Errorf("Expected SQL '%s', got '%s'", tc.expectedSQL, sql)
			}

			if !compareSlices(values, tc.expectedValues) {
				t.Errorf("Expected values %v, got %v", tc.expectedValues, values)
			}
		})
	}
}
