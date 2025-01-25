package sqlstmt

import (
	"testing"

	"github.com/insei/gerpo/types"
)

func Test_newValues(t *testing.T) {
	tests := []struct {
		name    string
		columns types.ExecutionColumns
	}{
		{
			name:    "Should create non-nil instance",
			columns: newMockExecutionColumns(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValues(tt.columns)
			if v == nil {
				t.Fatal("Expected non-nil values instance")
			}
			if v.columns != tt.columns {
				t.Error("Expected columns to match")
			}
		})
	}
}

func Test_WithModelValues(t *testing.T) {
	tests := []struct {
		name        string
		model       any
		mockValues  []any
		expectedLen int
	}{
		{
			name:        "Should keep empty when no model values",
			model:       nil,
			mockValues:  []any{},
			expectedLen: 0,
		},
		{
			name:        "Should add model values",
			model:       "some model",
			mockValues:  []any{"val1", "val2"},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCols := newMockExecutionColumns(nil)
			mockCols.modelValues = tt.mockValues

			v := newValues(mockCols)
			WithModelValues(tt.model).Apply(v)

			if len(v.values) != tt.expectedLen {
				t.Errorf("Expected %d values, got %d", tt.expectedLen, len(v.values))
			}
		})
	}
}
