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

func TestWithInsertProtection(t *testing.T) {
	testCases := []struct {
		name     string
		expected options
	}{
		{
			name:     "WithInsertProtection",
			expected: options{notAvailActions: []types.SQLAction{types.SQLActionInsert}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := options{}
			WithInsertProtection().apply(&opt)
			assert.Equal(t, tc.expected, opt)
		})
	}
}

func TestWithUpdateProtection(t *testing.T) {
	testCases := []struct {
		name     string
		expected options
	}{
		{
			name:     "WithUpdateProtection",
			expected: options{notAvailActions: []types.SQLAction{types.SQLActionUpdate}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := options{}
			WithUpdateProtection().apply(&opt)
			assert.Equal(t, tc.expected, opt)
		})
	}
}
