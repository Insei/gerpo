package column

import (
	"context"
	"testing"

	"github.com/insei/fmap/v3"

	"github.com/insei/gerpo/types"
)

type Test struct {
	Age  int
	Name string
}

func TestBuilderBuildWithAlias(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	field := fields.MustFind("Age")

	cases := []struct {
		name          string
		alias         string
		field         fmap.Field
		expectedAlias string
	}{
		{
			name:          "check alias 1",
			alias:         "test_alias",
			field:         field,
			expectedAlias: "age AS test_alias",
		},
		{
			name:          "check alias 2",
			alias:         "test",
			field:         field,
			expectedAlias: "age AS test",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewBuilder(tc.field)
			builder = builder.WithAlias(tc.alias)
			col := builder.Build()
			alias := col.ToSQL(context.Background())
			if alias != tc.expectedAlias {
				t.Errorf("expected alias to be %q, but got %q", tc.expectedAlias, alias)
			}
		})
	}
}

func TestBuilderBuildWithTable(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	fieldName := fields.MustFind("Name")
	fieldAge := fields.MustFind("Age")

	cases := []struct {
		name          string
		table         string
		field         fmap.Field
		expectedTable string
	}{
		{
			name:          "check table 1",
			table:         "user",
			field:         fieldName,
			expectedTable: "user.name",
		},
		{
			name:          "check table 2",
			table:         "student",
			field:         fieldName,
			expectedTable: "student.name",
		},
		{
			name:          "check empty table",
			table:         "",
			field:         fieldAge,
			expectedTable: "age",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewBuilder(tc.field)
			builder = builder.WithTable(tc.table)
			col := builder.Build()
			table := col.ToSQL(context.Background())
			if table != tc.expectedTable {
				t.Errorf("expected table to be %s, but got %s", tc.expectedTable, table)
			}
		})
	}
}

func TestBuilderBuildWithColumnName(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	fieldName := fields.MustFind("Name")
	fieldAge := fields.MustFind("Age")

	cases := []struct {
		name               string
		column             string
		field              fmap.Field
		expectedColumnName string
	}{
		{
			name:               "check column 1",
			column:             "test_column",
			field:              fieldName,
			expectedColumnName: "test_column",
		},
		{
			name:               "check column 2",
			column:             "another_column",
			field:              fieldName,
			expectedColumnName: "another_column",
		},
		{
			name:               "check empty column",
			column:             "",
			field:              fieldAge,
			expectedColumnName: "age",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewBuilder(tc.field)
			builder = builder.WithColumnName(tc.column)
			col := builder.Build()
			columnName := col.ToSQL(context.Background())
			if columnName != tc.expectedColumnName {
				t.Errorf("expected column to be %s, but got %s", tc.expectedColumnName, columnName)
			}
		})
	}
}

func TestBuilderBuildWithInsertProtection(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	fieldName := fields.MustFind("Name")
	fieldAge := fields.MustFind("Age")

	cases := []struct {
		name           string
		field          fmap.Field
		expectedAction types.SQLAction
	}{
		{
			name:           "check insert protection",
			field:          fieldName,
			expectedAction: types.SQLActionSelect,
		},
		{
			name:           "check insert protection on different field",
			field:          fieldAge,
			expectedAction: types.SQLActionSelect,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewBuilder(tc.field)
			builder = builder.WithInsertProtection()
			col := builder.Build()
			if col.IsAllowedAction(types.SQLActionInsert) {
				t.Errorf("expected insert action to be disallowed, but got allowed")
			}
			if !col.IsAllowedAction(tc.expectedAction) {
				t.Errorf("expected select action to be allowed, but got disallowed")
			}
		})
	}
}

func TestBuilderBuildWithUpdateProtection(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	fieldName := fields.MustFind("Name")
	fieldAge := fields.MustFind("Age")

	cases := []struct {
		name           string
		field          fmap.Field
		expectedAction types.SQLAction
	}{
		{
			name:           "check update protection",
			field:          fieldName,
			expectedAction: types.SQLActionSelect,
		},
		{
			name:           "check update protection on different field",
			field:          fieldAge,
			expectedAction: types.SQLActionSelect,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewBuilder(tc.field)
			builder = builder.WithUpdateProtection()
			col := builder.Build()
			if col.IsAllowedAction(types.SQLActionUpdate) {
				t.Errorf("expected update action to be disallowed, but got allowed")
			}
			if !col.IsAllowedAction(tc.expectedAction) {
				t.Errorf("expected select action to be allowed, but got disallowed")
			}
		})
	}
}
