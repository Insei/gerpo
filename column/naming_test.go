package column

import (
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"CamelCase", "camel_case"},
		{"PascalCase", "pascal_case"},
		{"snake_case", "snake_case"},
		{"mixedExample123", "mixed_example123"},
		{"ExampleWith123Numbers", "example_with123_numbers"},
		{"exampleWithMultipleCAPS", "example_with_multiple_caps"},
		{"JSONResponse", "json_response"},
		{"HTTPRequest", "http_request"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := toSnakeCase(tc.input)
			if got != tc.expected {
				t.Errorf("toSnakeCase(%s) = %s; expected %s", tc.input, got, tc.expected)
			}
		})
	}
}
