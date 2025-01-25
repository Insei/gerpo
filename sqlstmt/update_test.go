package sqlstmt

import (
	"context"
	"testing"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

func TestNewUpdate(t *testing.T) {
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
				&mockColumn{name: "id", hasName: true, allowedAction: true},
				&mockColumn{name: "name", hasName: true, allowedAction: true},
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
			u := NewUpdate(context.Background(), storage, tc.wantTable)

			if u.table != tc.wantTable {
				t.Errorf("Expected table %s, got %s", tc.wantTable, u.table)
			}
			if u.columns == nil {
				t.Error("Expected ExecutionColumns to be initialized")
			}
			if u.colsStorage != storage {
				t.Error("Expected colsStorage to match the provided storage")
			}
		})
	}
}

func TestUpdate_sql(t *testing.T) {
	tests := []struct {
		name             string
		table            string
		executionColumns []types.Column
		expectedSQL      string
	}{
		{
			name:  "BasicSQL",
			table: "users",
			executionColumns: []types.Column{
				&mockColumn{name: "id", hasName: true},
				&mockColumn{name: "name", hasName: true},
			},
			expectedSQL: "UPDATE users SET id = ?, name = ?",
		},
		{
			name:             "NoColumns",
			table:            "products",
			executionColumns: []types.Column{},
			expectedSQL:      "",
		},
		{
			name:  "SomeColumnsWithoutName",
			table: "orders",
			executionColumns: []types.Column{
				&mockColumn{name: "id", hasName: true},
				&mockColumn{name: "", hasName: false},
			},
			expectedSQL: "UPDATE orders SET id = ?",
		},
		{
			name:  "ColumnWithoutName",
			table: "orders",
			executionColumns: []types.Column{
				&mockColumn{name: "", hasName: false},
			},
			expectedSQL: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := newMockStorage(tc.executionColumns)
			u := NewUpdate(context.Background(), storage, tc.table)

			sqlStr := u.sql()

			if sqlStr != tc.expectedSQL {
				t.Errorf("Expected SQL '%s', got '%s'", tc.expectedSQL, sqlStr)
			}
		})
	}
}

// TestUpdate_SQL tests the SQL method of the Update struct.
func TestUpdate_SQL(t *testing.T) {
	tests := []struct {
		name             string
		table            string
		executionColumns []types.Column
		expectedSQL      string
		expectedValues   []any
	}{
		{
			name:  "BasicUpdate",
			table: "users",
			executionColumns: []types.Column{
				&mockColumn{name: "id", hasName: true, allowedAction: true},
				&mockColumn{name: "name", hasName: true, allowedAction: true},
			},
			expectedSQL:    "UPDATE users SET id = ?, name = ?",
			expectedValues: []any{},
		},
		{
			name:             "NoColumns",
			table:            "products",
			executionColumns: []types.Column{},
			expectedSQL:      "",
			expectedValues:   []any{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := newMockStorage(tc.executionColumns)
			ctx := context.Background()
			u := NewUpdate(ctx, storage, tc.table)

			u.where = sqlpart.NewWhereBuilder(ctx)

			sqlStr, vals := u.SQL()

			if sqlStr != tc.expectedSQL {
				t.Errorf("Expected SQL '%s', got '%s'", tc.expectedSQL, sqlStr)
			}
			if !compareSlices(vals, tc.expectedValues) {
				t.Errorf("Expected values %v, got %v", tc.expectedValues, vals)
			}
		})
	}
}
