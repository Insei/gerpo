package column

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insei/gerpo/types"
)

func TestWithAlias(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected options
	}{
		{
			name:     "With alias",
			input:    "aliasName",
			expected: options{alias: "aliasName"},
		},
		{
			name:     "With empty alias",
			input:    "",
			expected: options{alias: ""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := options{}
			WithAlias(tc.input).apply(&opt)
			assert.Equal(t, tc.expected, opt)
		})
	}
}

func TestWithTable(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected options
	}{
		{
			name:     "With table",
			input:    "tableName",
			expected: options{table: "tableName"},
		},
		{
			name:     "With empty table",
			input:    "",
			expected: options{table: ""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := options{}
			WithTable(tc.input).apply(&opt)
			assert.Equal(t, tc.expected, opt)
		})
	}
}

func TestWithColumnName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected options
	}{
		{
			name:     "WithLeadingAndTrailingSpaces",
			input:    " columnName ",
			expected: options{name: "columnName"},
		},
		{
			name:     "EmptyColumnName",
			input:    "",
			expected: options{name: ""},
		},
		{
			name:     "WithInternalSpaces",
			input:    "column Name",
			expected: options{name: "column Name"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := options{}
			WithColumnName(tc.input).apply(&opt)
			assert.Equal(t, tc.expected, opt)
		})
	}
}

func TestWithOmitOnInsert(t *testing.T) {
	testCases := []struct {
		name     string
		expected options
	}{
		{
			name:     "WithOmitOnInsert",
			expected: options{notAvailActions: []types.SQLAction{types.SQLActionInsert}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := options{}
			WithOmitOnInsert().apply(&opt)
			assert.Equal(t, tc.expected, opt)
		})
	}
}

func TestWithOmitOnUpdate(t *testing.T) {
	testCases := []struct {
		name     string
		expected options
	}{
		{
			name:     "WithOmitOnUpdate",
			expected: options{notAvailActions: []types.SQLAction{types.SQLActionUpdate}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := options{}
			WithOmitOnUpdate().apply(&opt)
			assert.Equal(t, tc.expected, opt)
		})
	}
}
