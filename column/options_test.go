package column

import (
	"testing"

	"github.com/insei/gerpo/types"
)

func TestWithAlias(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"test_alias", "test_alias"},
		{"", ""},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			o := &options{}
			WithAlias(tc.input).apply(o)

			if o.alias != tc.expected {
				t.Errorf("expected alias '%s', but got '%s'", tc.expected, o.alias)
			}
		})
	}
}

func TestWithTable(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"test_table", "test_table"},
		{"", ""},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			o := &options{}
			WithTable(tc.input).apply(o)

			if o.table != tc.expected {
				t.Errorf("expected table '%s', but got '%s'", tc.expected, o.table)
			}
		})
	}
}

func TestWithColumnName(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{" test_name ", "test_name"},
		{"   spaced_name   ", "spaced_name"},
		{"", ""},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			o := &options{}
			WithColumnName(tc.input).apply(o)

			if o.name != tc.expected {
				t.Errorf("expected name '%s', but got '%s'", tc.expected, o.name)
			}
		})
	}
}

func TestWithInsertProtection(t *testing.T) {
	o := &options{}
	WithInsertProtection().apply(o)

	expected := []types.SQLAction{types.SQLActionInsert}
	if len(o.notAvailActions) != len(expected) || o.notAvailActions[0] != expected[0] {
		t.Errorf("expected notAvailActions %v, but got %v", expected, o.notAvailActions)
	}
}

func TestWithUpdateProtection(t *testing.T) {
	o := &options{}
	WithUpdateProtection().apply(o)

	expected := []types.SQLAction{types.SQLActionUpdate}
	if len(o.notAvailActions) != len(expected) || o.notAvailActions[0] != expected[0] {
		t.Errorf("expected notAvailActions %v, but got %v", expected, o.notAvailActions)
	}
}

func TestMultipleOptions(t *testing.T) {
	cases := []struct {
		name     string
		opts     []Option
		expected options
	}{
		{
			name: "multiple options",
			opts: []Option{
				WithAlias("mult_alias"),
				WithTable("mult_table"),
				WithColumnName("mult_name"),
				WithInsertProtection(),
				WithUpdateProtection(),
			},
			expected: options{
				alias:           "mult_alias",
				table:           "mult_table",
				name:            "mult_name",
				notAvailActions: []types.SQLAction{types.SQLActionInsert, types.SQLActionUpdate},
			},
		},
		{
			name: "no alias",
			opts: []Option{
				WithTable("no_alias_table"),
				WithColumnName("no_alias_name"),
				WithInsertProtection(),
			},
			expected: options{
				table:           "no_alias_table",
				name:            "no_alias_name",
				notAvailActions: []types.SQLAction{types.SQLActionInsert},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := &options{}
			for _, opt := range tc.opts {
				opt.apply(o)
			}

			if o.alias != tc.expected.alias {
				t.Errorf("expected alias '%s', but got '%s'", tc.expected.alias, o.alias)
			}
			if o.table != tc.expected.table {
				t.Errorf("expected table '%s', but got '%s'", tc.expected.table, o.table)
			}
			if o.name != tc.expected.name {
				t.Errorf("expected name '%s', but got '%s'", tc.expected.name, o.name)
			}
			if len(o.notAvailActions) != len(tc.expected.notAvailActions) {
				t.Errorf("expected %d notAvailActions, but got %d", len(tc.expected.notAvailActions), len(o.notAvailActions))
			}
			for i, action := range o.notAvailActions {
				if action != tc.expected.notAvailActions[i] {
					t.Errorf("expected notAvailAction '%s', but got '%s'", tc.expected.notAvailActions[i], action)
				}
			}
		})
	}
}
