package sqlstmt

import (
	"context"
	"testing"

	"github.com/insei/gerpo/types"
	"github.com/stretchr/testify/assert"
)

func TestNewDelete(t *testing.T) {
	testCases := []struct {
		name     string
		table    string
		storage  types.ColumnsStorage
		expected string
	}{
		{
			name:     "Initialize Delete with valid table",
			table:    "users",
			storage:  newMockStorage(nil),
			expected: "users",
		},
		{
			name:     "Initialize Delete with empty table",
			table:    "",
			storage:  newMockStorage(nil),
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			deleteStmt := NewDelete(ctx, tc.table, tc.storage)

			if deleteStmt.table != tc.expected {
				t.Errorf("Expected table '%s', got '%s'", tc.expected, deleteStmt.table)
			}

			if deleteStmt.ctx != ctx {
				t.Errorf("Expected context '%v', got '%v'", ctx, deleteStmt.ctx)
			}

			if deleteStmt.columnsStorage != tc.storage {
				t.Errorf("Expected ColumnsStorage '%v', got '%v'", tc.storage, deleteStmt.columnsStorage)
			}
		})
	}
}

func TestDelete_SQL(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(deleteStmt *Delete)
		expectedSQL    string
		expectedValues []any
		expectError    bool
	}{
		{
			name: "Generate SQL without conditions",
			setup: func(deleteStmt *Delete) {
				// No additional setup
			},
			expectedSQL:    "DELETE FROM users",
			expectedValues: []any{},
		},
		{
			name: "Generate SQL with WHERE condition",
			setup: func(deleteStmt *Delete) {
				deleteStmt.where.AppendSQLWithValues("age > ?", true, 30)
			},
			expectedSQL:    "DELETE FROM users WHERE age > ?",
			expectedValues: []any{30},
		},
		{
			name: "Generate SQL with multiple WHERE conditions",
			setup: func(deleteStmt *Delete) {
				deleteStmt.where.AppendSQLWithValues("age > ?", true, 25)
				deleteStmt.where.AND()
				deleteStmt.where.AppendSQLWithValues("city = ?", true, "New York")
			},
			expectedSQL:    "DELETE FROM users WHERE age > ? AND city = ?",
			expectedValues: []any{25, "New York"},
		},
		{
			name: "Generate SQL with JOIN",
			setup: func(deleteStmt *Delete) {
				deleteStmt.join.JOIN(func(ctx context.Context) string {
					return "INNER JOIN orders ON users.id = orders.user_id"
				})
			},
			expectedSQL:    "DELETE FROM users INNER JOIN orders ON users.id = orders.user_id",
			expectedValues: []any{},
		},
		{
			name: "Generate SQL with WHERE and JOIN",
			setup: func(deleteStmt *Delete) {
				deleteStmt.join.JOIN(func(ctx context.Context) string {
					return "INNER JOIN orders ON users.id = orders.user_id"
				})
				deleteStmt.where.AppendSQLWithValues("orders.amount > ?", true, 100)
			},
			expectedSQL:    "DELETE FROM users INNER JOIN orders ON users.id = orders.user_id WHERE orders.amount > ?",
			expectedValues: []any{100},
		},
		{
			name: "Generate SQL with empty table",
			setup: func(deleteStmt *Delete) {
				deleteStmt.table = ""
			},
			expectedSQL:    "",
			expectedValues: []any{},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			deleteStmt := NewDelete(ctx, "users", nil)

			tc.setup(deleteStmt)

			sql, values, err := deleteStmt.SQL()
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
