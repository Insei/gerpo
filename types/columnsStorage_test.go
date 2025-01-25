package types

import (
	"context"
	"testing"

	"github.com/insei/fmap/v3"
)

type mockColumn struct {
	Column
	name          string
	allowedAction bool
	hasName       bool

	field fmap.Field
}

func (m *mockColumn) IsAllowedAction(action SQLAction) bool {
	return m.allowedAction
}

func (m *mockColumn) ToSQL(ctx context.Context) string {
	return m.name
}

func (m *mockColumn) Name() (string, bool) {
	return m.name, m.hasName
}

func (m *mockColumn) GetField() fmap.Field {
	return m.field
}

type mockField struct {
	fmap.Field
}

func Test_columnsStorage_Add(t *testing.T) {
	tests := []struct {
		name          string
		mockCols      []Column
		checkSQLRoles []SQLAction
	}{
		{
			name:          "Add single column",
			mockCols:      []Column{&mockColumn{name: "col1", allowedAction: true, field: &mockField{}}},
			checkSQLRoles: []SQLAction{SQLActionInsert, SQLActionSelect, SQLActionUpdate, SQLActionGroup, SQLActionSort},
		},
		{
			name: "Add multiple columns",
			mockCols: []Column{
				&mockColumn{name: "col2", allowedAction: true, field: &mockField{}},
				&mockColumn{name: "col3", allowedAction: false, field: &mockField{}},
			},
			checkSQLRoles: []SQLAction{SQLActionInsert, SQLActionSelect, SQLActionUpdate, SQLActionGroup, SQLActionSort},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewEmptyColumnsStorage(nil)

			for _, col := range tt.mockCols {
				storage.Add(col)
			}

			slice := storage.AsSlice()
			if len(slice) != len(tt.mockCols) {
				t.Errorf("Expected %d columns, got %d", len(tt.mockCols), len(slice))
			}

			for _, col := range tt.mockCols {
				found := false
				for _, sCol := range slice {
					if sCol == col {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected column %v not found in storage slice", col)
				}
			}
		})
	}
}

func Test_columnsStorage_NewExecutionColumns(t *testing.T) {
	tests := []struct {
		name         string
		mockCols     []Column
		allowed      bool
		action       SQLAction
		expectedCols int
	}{
		{
			name:         "No columns for action",
			mockCols:     []Column{},
			allowed:      false,
			action:       SQLActionInsert,
			expectedCols: 0,
		},
		{
			name: "Two columns allowed for action",
			mockCols: []Column{
				&mockColumn{name: "col4", allowedAction: true},
				&mockColumn{name: "col5", allowedAction: true},
			},
			allowed:      true,
			action:       SQLActionInsert,
			expectedCols: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewEmptyColumnsStorage(nil)
			for _, col := range tt.mockCols {
				storage.Add(col)
			}

			execCols := storage.NewExecutionColumns(context.Background(), tt.action)
			if tt.expectedCols == 0 && execCols != nil {
				t.Error("Expected nil ExecutionColumns for no columns added")
				return
			}
			if tt.expectedCols > 0 && execCols == nil {
				t.Error("Expected non-nil ExecutionColumns when columns are allowed")
				return
			}

			if execCols != nil {
				all := execCols.GetAll()
				if len(all) != tt.expectedCols {
					t.Errorf("Expected %d columns, got %d", tt.expectedCols, len(all))
				}
			}
		})
	}
}
