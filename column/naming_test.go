package column

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSnakeCase(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Single word",
			input:    "Test",
			expected: "test",
		},
		{
			name:     "CamelCase",
			input:    "CamelCase",
			expected: "camel_case",
		},
		{
			name:     "camelCase with initial lower",
			input:    "camelCase",
			expected: "camel_case",
		},
		{
			name:     "MixedCaps",
			input:    "MixedCAPSAndCamel",
			expected: "mixed_caps_and_camel",
		},
		{
			name:     "MultipleWords",
			input:    "ThisIsATest",
			expected: "this_is_a_test",
		},
		{
			name:     "LongWords",
			input:    "ConvertThisToSnakeCase",
			expected: "convert_this_to_snake_case",
		},
		{
			name:     "AlreadySnakeCase",
			input:    "already_snake_case",
			expected: "already_snake_case",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := toSnakeCase(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
