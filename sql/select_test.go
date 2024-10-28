package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insei/gerpo/types"
)

func TestStringSelectBuilderExclude(t *testing.T) {
	col1 := &testColumn{allowedActions: map[types.SQLAction]bool{types.SQLActionSelect: true}}
	col2 := &testColumn{allowedActions: map[types.SQLAction]bool{types.SQLActionSelect: false}}
	builder := &StringSelectBuilder{
		columns: []types.Column{col1, col2},
	}

	testCases := []struct {
		cols     []types.Column
		expected []types.Column
	}{
		{
			cols:     []types.Column{col1},
			expected: []types.Column{col2},
		},
		{
			cols:     []types.Column{col1, col2},
			expected: []types.Column{},
		},
	}

	for _, tc := range testCases {
		builder.Exclude(tc.cols...)
		assert.Equal(t, tc.expected, builder.columns)
	}
}

func TestStringSelectBuilderLimit(t *testing.T) {
	builder := &StringSelectBuilder{}

	testCases := []struct {
		limit    uint64
		expected uint64
	}{
		{
			limit:    10,
			expected: 10,
		},
		{
			limit:    0,
			expected: 0,
		},
	}

	for _, tc := range testCases {
		builder.Limit(tc.limit)
		assert.Equal(t, tc.expected, builder.limit)
	}
}

func TestStringSelectBuilderOffset(t *testing.T) {
	builder := &StringSelectBuilder{}

	testCases := []struct {
		offset   uint64
		expected uint64
	}{
		{
			offset:   20,
			expected: 20,
		},
		{
			offset:   0,
			expected: 0,
		},
	}

	for _, tc := range testCases {
		builder.Offset(tc.offset)
		assert.Equal(t, tc.expected, builder.offset)
	}
}

func TestStringSelectBuilderOrderBy(t *testing.T) {
	builder := &StringSelectBuilder{}

	testCases := []struct {
		columnDirection string
		expected        string
	}{
		{
			columnDirection: "col1 ASC",
			expected:        "col1 ASC",
		},
		{
			columnDirection: "col2 DESC",
			expected:        "col1 ASC, col2 DESC",
		},
	}

	for _, tc := range testCases {
		builder.OrderBy(tc.columnDirection)
		assert.Equal(t, tc.expected, builder.orderBy)
	}
}

func TestStringSelectBuilderOrderByColumn(t *testing.T) {
	col1 := &testColumn{allowedActions: map[types.SQLAction]bool{types.SQLActionSort: true}}
	col2 := &testColumn{allowedActions: map[types.SQLAction]bool{types.SQLActionSort: false}}
	builder := &StringSelectBuilder{
		orderBy: "col ASC",
	}

	testCases := []struct {
		col types.Column
		dir types.OrderDirection
	}{
		{
			col: col1,
			dir: types.OrderDirectionDESC,
		},
		{
			col: col2,
			dir: types.OrderDirectionASC,
		},
	}

	for _, tc := range testCases {
		err := builder.OrderByColumn(tc.col, tc.dir)
		assert.NoError(t, err)
	}
}

func TestStringSelectBuilderGetColumns(t *testing.T) {
	col1 := &testColumn{}
	col2 := &testColumn{}
	builder := &StringSelectBuilder{
		columns: []types.Column{col1, col2},
	}

	expectedColumns := []types.Column{col1, col2}
	assert.Equal(t, expectedColumns, builder.GetColumns())
}

func TestStringSelectBuilderGetSQL(t *testing.T) {
	col1 := &testColumn{sql: "col1"}
	col2 := &testColumn{sql: "col2"}
	builder := &StringSelectBuilder{
		columns: []types.Column{col1, col2},
	}

	expectedSQL := "col1, col2"
	assert.Equal(t, expectedSQL, builder.GetSQL())
}

func TestStringSelectBuilderGetLimit(t *testing.T) {
	builder := &StringSelectBuilder{}

	testCases := []struct {
		name     string
		limit    uint64
		expected string
	}{
		{
			name:     "Limit is 10",
			limit:    10,
			expected: "10",
		},
		{
			name:     "Limit is 0",
			limit:    0,
			expected: "",
		},
		{
			name:     "Limit is 100",
			limit:    100,
			expected: "100",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder.limit = tc.limit
			assert.Equal(t, tc.expected, builder.GetLimit())
		})
	}
}

func TestStringSelectBuilderGetOffset(t *testing.T) {
	builder := &StringSelectBuilder{}

	testCases := []struct {
		name     string
		offset   uint64
		expected string
	}{
		{
			name:     "Offset is 1",
			offset:   1,
			expected: "1",
		},
		{
			name:     "Offset is 0",
			offset:   0,
			expected: "",
		},
		{
			name:     "Offset is 100",
			offset:   10,
			expected: "10",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder.offset = tc.offset
			assert.Equal(t, tc.expected, builder.GetOffset())
		})
	}
}

func TestStringSelectBuilderGetOrderSQL(t *testing.T) {
	builder := &StringSelectBuilder{
		orderBy: "col1 ASC, col2 DESC",
	}

	expectedSQL := "col1 ASC, col2 DESC"
	assert.Equal(t, expectedSQL, builder.GetOrderSQL())
}

func TestDeleteFunc(t *testing.T) {
	testCases := []struct {
		name     string
		input    []int
		toDelete func(int) bool
		expected []int
	}{
		{
			name:     "Delete first occurrence",
			input:    []int{1, 2, 3, 4, 5},
			toDelete: func(v int) bool { return v == 3 },
			expected: []int{4, 5},
		},
		{
			name:     "Delete multiple occurrences",
			input:    []int{1, 2, 3, 4, 5, 3, 6},
			toDelete: func(v int) bool { return v == 3 },
			expected: []int{4, 5, 6},
		},
		{
			name:     "No matches",
			input:    []int{1, 2, 3, 4, 5},
			toDelete: func(v int) bool { return v == 10 },
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Empty slice",
			input:    []int{},
			toDelete: func(v int) bool { return v == 3 },
			expected: []int{},
		},
		{
			name:     "Delete all elements",
			input:    []int{3, 3, 3},
			toDelete: func(v int) bool { return v == 3 },
			expected: []int{},
		},
		{
			name:     "Delete last element",
			input:    []int{1, 2, 3, 4, 5},
			toDelete: func(v int) bool { return v == 5 },
			expected: []int{},
		},
		{
			name:     "Delete first element",
			input:    []int{1, 2, 3, 4, 5},
			toDelete: func(v int) bool { return v == 1 },
			expected: []int{2, 3, 4, 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := deleteFunc(tc.input, tc.toDelete)
			assert.Equal(t, tc.expected, result)
		})
	}
}
