package sqlstmt

import (
	"context"
	"testing"

	"github.com/insei/gerpo/types"
)

// TestNewGetList tests the NewGetList constructor using a mock for columns storage.
func TestNewGetList(t *testing.T) {
	tests := []struct {
		name             string
		table            string
		executionColumns []types.Column
		wantTable        string
	}{
		{
			name:  "ValidTable",
			table: "users",
			executionColumns: []types.Column{
				&mockColumn{name: "id", hasName: true},
				&mockColumn{name: "name", hasName: true},
			},
			wantTable: "users",
		},
		{
			name:             "EmptyTable",
			table:            "",
			executionColumns: []types.Column{},
			wantTable:        "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := newMockStorage(tc.executionColumns)
			gl := NewGetList(context.Background(), tc.wantTable, storage)
			if gl.table != tc.wantTable {
				t.Errorf("Expected table %s, got %s", tc.wantTable, gl.table)
			}
			if gl.limitOffset == nil {
				t.Error("Expected LimitOffsetBuilder to be initialized")
			}
			if gl.columns == nil {
				t.Error("Expected ExecutionColumns to be initialized")
			}
		})
	}
}

// TestGetList_SQL tests the SQL method of the GetList struct using NewGetList.
func TestGetList_SQL(t *testing.T) {
	tests := []struct {
		name             string
		table            string
		executionColumns []types.Column
		expectedSQL      string
		expectedValues   []any
	}{
		{
			name:  "BasicSelect",
			table: "users",
			executionColumns: []types.Column{
				&mockColumn{name: "id", hasName: true},
				&mockColumn{name: "name", hasName: true},
			},
			expectedSQL:    "SELECT id, name FROM users",
			expectedValues: []any{},
		},
		{
			name:             "EmptyColumns",
			table:            "products",
			executionColumns: []types.Column{},
			expectedSQL:      "",
			expectedValues:   []any{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := newMockStorage(tc.executionColumns)
			gl := NewGetList(context.Background(), tc.table, storage)
			sql, values := gl.SQL()
			if sql != tc.expectedSQL {
				t.Errorf("Expected SQL '%s', got '%s'", tc.expectedSQL, sql)
			}
			if !compareSlices(values, tc.expectedValues) {
				t.Errorf("Expected values %v, got %v", tc.expectedValues, values)
			}
		})
	}
}
