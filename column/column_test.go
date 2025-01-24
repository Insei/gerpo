package column

import (
	"context"
	"testing"

	"github.com/insei/fmap/v3"
	"github.com/stretchr/testify/assert"

	"github.com/insei/gerpo/types"
)

func TestColumnGetFilterFn(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")
	c := column{
		base: &types.ColumnBase{
			Filters: types.NewFilterManagerForField(field),
		},
	}

	tests := []struct {
		name      string
		operation types.Operation
	}{
		{"ValidOperationEQ", types.OperationEQ},
		{"ValidOperationNEQ", types.OperationNEQ},
		{"InvalidOperation", types.Operation("INVALID")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.GetFilterFn(tt.operation)
		})
	}
}

func TestColumnIsAllowedAction(t *testing.T) {
	c := column{
		base: &types.ColumnBase{
			AllowedActions: []types.SQLAction{types.SQLActionSelect, types.SQLActionUpdate},
		},
	}

	tests := []struct {
		action   types.SQLAction
		expected bool
	}{
		{types.SQLActionInsert, false},
		{types.SQLActionSelect, true},
		{types.SQLActionUpdate, true},
	}

	for _, tt := range tests {
		result := c.IsAllowedAction(tt.action)
		assert.Equal(t, tt.expected, result)
	}
}

func TestColumnToSQL(t *testing.T) {
	c := column{
		base: &types.ColumnBase{
			ToSQL: func(ctx context.Context) string {
				return "SELECT * FROM table"
			},
		},
	}

	tests := []struct {
		ctx context.Context
	}{
		{context.Background()},
	}

	for _, tt := range tests {
		c.ToSQL(tt.ctx)
	}
}

func TestColumnGetPtr(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")
	c := column{
		base: &types.ColumnBase{
			GetPtr: func(model any) any {
				return field.GetPtr(model)
			},
		},
	}

	model := &TestModel{Age: 30}
	ptr := c.GetPtr(model)

	assert.NotNil(t, ptr)
}

func TestColumnGetField(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")
	c := column{
		base: &types.ColumnBase{
			Field: field,
		},
	}

	field2 := c.GetField()

	if field2 != field {
		t.Errorf("Expected field %v, got %v", field, field2)
	}
}

func TestColumnName(t *testing.T) {
	c := column{
		name: "test",
	}

	name, _ := c.Name()

	assert.Equal(t, c.name, name)
}

func TestColumnTable(t *testing.T) {
	c := column{
		table: "test",
	}

	table, _ := c.Table()

	assert.Equal(t, c.table, table)
}

func TestColumnGetAllowedActions(t *testing.T) {
	c := column{
		base: &types.ColumnBase{
			AllowedActions: []types.SQLAction{types.SQLActionInsert, types.SQLActionSelect},
		},
	}

	actions := c.GetAllowedActions()

	assert.NotNil(t, actions)
}

func TestColumnGetAvailableFilterOperations(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")
	c := column{
		base: &types.ColumnBase{
			Filters: types.NewFilterManagerForField(field),
		},
	}

	c.GetAvailableFilterOperations()
}

func TestColumnIsAvailableFilterOperation(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")
	c := column{
		base: &types.ColumnBase{
			Filters: types.NewFilterManagerForField(field),
		},
	}

	c.IsAvailableFilterOperation(types.OperationEQ)
}

func TestGenerateSQLQuery(t *testing.T) {
	tests := []struct {
		name     string
		options  options
		expected string
	}{
		{
			name:     "No table",
			options:  options{name: "columnName"},
			expected: "columnName",
		},
		{
			name:     "With table",
			options:  options{table: "table", name: "columnName"},
			expected: "table.columnName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSQLColumnString(&tt.options)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateToSQLFn(t *testing.T) {
	tests := []struct {
		name   string
		sql    string
		alias  string
		ctx    context.Context
		expect string
	}{
		{
			name:   "No alias",
			sql:    "SELECT * FROM table",
			alias:  "",
			ctx:    context.Background(),
			expect: "SELECT * FROM table",
		},
		{
			name:   "With alias",
			sql:    "SELECT * FROM table",
			alias:  "t",
			ctx:    context.Background(),
			expect: "SELECT * FROM table AS t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toSQLFn := generateToSQLFn(tt.sql, tt.alias)
			result := toSQLFn(tt.ctx)
			assert.Equal(t, tt.expect, result)
		})
	}
}
