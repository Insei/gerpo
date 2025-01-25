package sqlstmt

import (
	"context"
	"testing"

	"github.com/insei/gerpo/types"
)

func TestNewCount(t *testing.T) {
	testCases := []struct {
		name     string
		table    string
		storage  types.ColumnsStorage
		expected string
	}{
		{
			name:     "Initialize Count with valid table",
			table:    "users",
			storage:  newMockStorage(nil),
			expected: "users",
		},
		{
			name:     "Initialize Count with empty table",
			table:    "",
			storage:  newMockStorage(nil),
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			count := NewCount(ctx, tc.table, tc.storage)

			if count.table != tc.expected {
				t.Errorf("Expected table '%s', got '%s'", tc.expected, count.table)
			}
		})
	}
}

func TestCount_SQL(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(count *Count)
		expectedSQL    string
		expectedValues []any
	}{
		{
			name: "Generate SQL without conditions",
			setup: func(count *Count) {
				// Нет дополнительных настроек
			},
			expectedSQL:    "SELECT count(*) over() AS count FROM users LIMIT 1",
			expectedValues: []any{},
		},
		{
			name: "Generate SQL with WHERE condition",
			setup: func(count *Count) {
				count.where.AppendSQLWithValues("age > ?", true, 30)
			},
			expectedSQL:    "SELECT count(*) over() AS count FROM users WHERE age > ? LIMIT 1",
			expectedValues: []any{30},
		},
		{
			name: "Generate SQL with multiple WHERE conditions",
			setup: func(count *Count) {
				count.where.AppendSQLWithValues("age > ?", true, 25)
				count.where.AND()
				count.where.AppendSQLWithValues("city = ?", true, "New York")
			},
			expectedSQL:    "SELECT count(*) over() AS count FROM users WHERE age > ? AND city = ? LIMIT 1",
			expectedValues: []any{25, "New York"},
		},
		{
			name: "Generate SQL with GROUP BY",
			setup: func(count *Count) {
				count.group.GroupBy(&mockColumn{name: "department", allowedAction: true})
			},
			expectedSQL:    "SELECT count(*) over() AS count FROM users GROUP BY department LIMIT 1",
			expectedValues: []any{},
		},
		{
			name: "Generate SQL with ORDER BY",
			setup: func(count *Count) {
				count.order.OrderByColumn(&mockColumn{name: "created_at", allowedAction: true}, "DESC")
			},
			expectedSQL:    "SELECT count(*) over() AS count FROM users LIMIT 1",
			expectedValues: []any{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			count := NewCount(ctx, "users", nil)

			tc.setup(count)

			sql, values := count.SQL()

			if sql != tc.expectedSQL {
				t.Errorf("Expected SQL '%s', got '%s'", tc.expectedSQL, sql)
			}

			if !compareSlices(values, tc.expectedValues) {
				t.Errorf("Expected values %v, got %v", tc.expectedValues, values)
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
