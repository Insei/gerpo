package sqlpart

import (
	"context"
	"testing"

	"github.com/insei/gerpo/types"
)

// MockColumn implements the types.Column interface for testing
type MockColumn struct {
	types.Column
	name          string
	allowedAction bool
}

func (m *MockColumn) IsAllowedAction(action types.SQLAction) bool {
	return m.allowedAction
}

func (m *MockColumn) ToSQL(ctx context.Context) string {
	return m.name
}

// TestOrderBuilder_OrderBy tests the OrderBy method using test cases
func TestOrderBuilder_OrderBy(t *testing.T) {
	testCases := []struct {
		name     string
		actions  []string
		expected string
	}{
		{
			name:     "Add single order condition",
			actions:  []string{"name ASC"},
			expected: "name ASC",
		},
		{
			name:     "Add multiple order conditions",
			actions:  []string{"name ASC", "age DESC"},
			expected: "name ASC, age DESC",
		},
		{
			name:     "No order conditions",
			actions:  []string{},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			builder := NewOrderBuilder(ctx)

			for _, action := range tc.actions {
				builder.OrderBy(action)
			}

			if builder.orderBy != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, builder.orderBy)
			}
		})
	}
}

// TestOrderBuilder_OrderByColumn tests the OrderByColumn method using test cases
func TestOrderBuilder_OrderByColumn(t *testing.T) {
	testCases := []struct {
		name          string
		column        types.Column
		direction     types.OrderDirection
		initialOrder  string
		expectedOrder string
	}{
		{
			name:          "Add allowed column",
			column:        &MockColumn{name: "created_at", allowedAction: true},
			direction:     "DESC",
			initialOrder:  "",
			expectedOrder: "created_at DESC",
		},
		{
			name:          "Add second allowed column",
			column:        &MockColumn{name: "updated_at", allowedAction: true},
			direction:     "ASC",
			initialOrder:  "created_at DESC",
			expectedOrder: "created_at DESC, updated_at ASC",
		},
		{
			name:          "Attempt to add disallowed column",
			column:        &MockColumn{name: "password", allowedAction: false},
			direction:     "ASC",
			initialOrder:  "created_at DESC",
			expectedOrder: "created_at DESC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			builder := NewOrderBuilder(ctx)
			builder.orderBy = tc.initialOrder

			builder.OrderByColumn(tc.column, tc.direction)

			if builder.orderBy != tc.expectedOrder {
				t.Errorf("Expected '%s', got '%s'", tc.expectedOrder, builder.orderBy)
			}
		})
	}
}

// TestOrderBuilder_SQL tests the SQL method using test cases
func TestOrderBuilder_SQL(t *testing.T) {
	testCases := []struct {
		name        string
		actions     []string
		expectedSQL string
	}{
		{
			name:        "Empty SQL query",
			actions:     []string{},
			expectedSQL: "",
		},
		{
			name:        "Single order condition",
			actions:     []string{"name ASC"},
			expectedSQL: " ORDER BY name ASC",
		},
		{
			name:        "Multiple order conditions",
			actions:     []string{"name ASC", "age DESC"},
			expectedSQL: " ORDER BY name ASC, age DESC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			builder := NewOrderBuilder(ctx)

			for _, action := range tc.actions {
				builder.OrderBy(action)
			}

			sql := builder.SQL()
			if sql != tc.expectedSQL {
				t.Errorf("Expected '%s', got '%s'", tc.expectedSQL, sql)
			}
		})
	}
}
