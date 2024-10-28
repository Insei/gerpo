package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insei/gerpo/types"
)

func TestStringUpdateBuilderExclude(t *testing.T) {
	col1 := &testColumn{sql: "col1"}
	col2 := &testColumn{sql: "col2"}
	col3 := &testColumn{sql: "col3"}
	builder := &StringUpdateBuilder{
		columns: []types.Column{col1, col2, col3},
	}

	testCases := []struct {
		name            string
		colsToExclude   []types.Column
		expectedColumns []types.Column
	}{
		{
			name:            "Exclude single column",
			colsToExclude:   []types.Column{col1},
			expectedColumns: []types.Column{col2, col3},
		},
		{
			name:            "Exclude multiple columns",
			colsToExclude:   []types.Column{col1, col2},
			expectedColumns: []types.Column{col3},
		},
		{
			name:            "Exclude all columns",
			colsToExclude:   []types.Column{col1, col2, col3},
			expectedColumns: []types.Column{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder.Exclude(tc.colsToExclude...)
			assert.Equal(t, tc.expectedColumns, builder.GetColumns())
		})
	}
}

func TestStringUpdateBuilderGetColumns(t *testing.T) {
	col1 := &testColumn{sql: "col1"}
	col2 := &testColumn{sql: "col2"}
	col3 := &testColumn{sql: "col3"}
	builder := &StringUpdateBuilder{
		columns: []types.Column{col1, col2, col3},
	}

	expectedColumns := []types.Column{col1, col2, col3}
	assert.Equal(t, expectedColumns, builder.GetColumns())
}

func TestStringUpdateBuilderSQL(t *testing.T) {
	col1 := &testColumn{sql: "col1"}
	col2 := &testColumn{sql: "col2"}
	col3 := &testColumn{sql: "col3"}
	col4 := &testColumn{}

	testCases := []struct {
		name        string
		columns     []types.Column
		expectedSQL string
	}{
		{
			name:        "All columns",
			columns:     []types.Column{col1, col2, col3},
			expectedSQL: "col1 = ?, col2 = ?, col3 = ?",
		},
		{
			name:        "No columns",
			columns:     []types.Column{},
			expectedSQL: "",
		},
		{
			name:        "Columns with empty name",
			columns:     []types.Column{col1, col4, col3},
			expectedSQL: "col1 = ?, col3 = ?",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := &StringUpdateBuilder{
				columns: tc.columns,
			}
			sql := builder.SQL()
			assert.Equal(t, tc.expectedSQL, sql)
		})
	}
}
