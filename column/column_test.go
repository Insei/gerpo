package column

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/insei/fmap/v3"

	"github.com/insei/gerpo/types"
)

func TestColumnGetFilterFn(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	tests := []struct {
		fieldName string
		operation types.Operation
		value     any
		expected  string
	}{
		{
			fieldName: "Age",
			operation: types.OperationGT,
			value:     25,
			expected:  "test_table.age > ?",
		},
		{
			fieldName: "Name",
			operation: types.OperationEQ,
			value:     "John",
			expected:  "test_table.name = ?",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("field: %s, operation: %s", test.fieldName, test.operation), func(t *testing.T) {
			field := fields.MustFind(test.fieldName)
			col := New(field, []Option{WithTable("test_table")}...)

			filterFn, ok := col.GetFilterFn(test.operation)
			if !ok {
				t.Errorf("expected true, but got false")
			}

			result, success, err := filterFn(context.Background(), test.value)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !success {
				t.Errorf("expected true, but got false")
			}

			if result != test.expected {
				t.Errorf("expected result %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestColumnIsAllowedAction(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	tests := []struct {
		fieldName      string
		allowedActions []types.SQLAction
		action         types.SQLAction
		expected       bool
	}{
		{
			fieldName:      "Age",
			allowedActions: []types.SQLAction{types.SQLActionSelect, types.SQLActionInsert},
			action:         types.SQLActionUpdate,
			expected:       true,
		},
		{
			fieldName:      "Name",
			allowedActions: []types.SQLAction{types.SQLActionUpdate},
			action:         types.SQLActionInsert,
			expected:       false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("field: %s, action: %s", test.fieldName, test.action), func(t *testing.T) {
			field := fields.MustFind(test.fieldName)
			col := New(field, []Option{WithTable("test_table"), WithInsertProtection()}...)

			result := col.IsAllowedAction(test.action)
			if result != test.expected {
				t.Errorf("expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestColumnToSQL(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	tests := []struct {
		fieldName   string
		tableName   string
		expectedSQL string
	}{
		{
			fieldName:   "Age",
			tableName:   "student",
			expectedSQL: "student.age",
		},
		{
			fieldName:   "Name",
			tableName:   "worker",
			expectedSQL: "worker.name",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("field: %s", test.fieldName), func(t *testing.T) {
			field := fields.MustFind(test.fieldName)
			col := New(field, []Option{WithTable(test.tableName)}...)

			sql := col.ToSQL(context.Background())
			if sql != test.expectedSQL {
				t.Errorf("expected %v, but got %v", test.expectedSQL, sql)
			}
		})
	}
}

func TestColumnGetAvailableFilterOperations(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	field := fields.MustFind("Age")
	col := New(field)

	operations := col.GetAvailableFilterOperations()
	expectedOperations := []types.Operation{types.OperationIN, types.OperationNIN, types.OperationEQ, types.OperationNEQ,
		types.OperationLT, types.OperationLTE, types.OperationGT, types.OperationGTE}

	if len(operations) != len(expectedOperations) {
		t.Fatalf("expected %d filter operations, but got %d", len(expectedOperations), len(operations))
	}

	for _, expectedOp := range expectedOperations {
		if !slices.Contains(operations, expectedOp) {
			t.Errorf("expected operation '%s' not found in available operations", expectedOp)
		}
	}

	for _, op := range operations {
		if !slices.Contains(expectedOperations, op) {
			t.Errorf("unexpected operation '%s' found in available operations", op)
		}
	}
}

func TestColumnIsAvailableFilterOperation(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	field := fields.MustFind("Age")
	col := New(field)

	if !col.IsAvailableFilterOperation(types.OperationIN) {
		t.Errorf("expected filter operation 'in' to be available, but it was not")
	}

	if col.IsAvailableFilterOperation("nonexistent_op") {
		t.Errorf("expected filter operation 'nonexistent_op' to be unavailable, but it was available")
	}
}

func TestColumnGetPtr(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	tests := []struct {
		fieldName  string
		initialVal any
		expected   any
	}{
		{
			fieldName:  "Age",
			initialVal: &Test{Age: 25},
			expected:   25,
		},
		{
			fieldName:  "Name",
			initialVal: &Test{Name: "GoGoland"},
			expected:   "GoGoland",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("field: %s", test.fieldName), func(t *testing.T) {
			field := fields.MustFind(test.fieldName)
			col := New(field)

			ptr := col.GetPtr(test.initialVal)
			if reflect.ValueOf(ptr).Elem().Interface() != test.expected {
				t.Errorf("expected value %v, but got %v", test.expected, reflect.ValueOf(ptr).Elem().Interface())
			}
		})
	}
}

func TestColumnGetField(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	tests := []struct {
		fieldName string
	}{
		{
			fieldName: "Age",
		},
		{
			fieldName: "Name",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("field: %s", test.fieldName), func(t *testing.T) {
			field := fields.MustFind(test.fieldName)
			col := New(field)

			returnedField := col.GetField()
			if !reflect.DeepEqual(returnedField, field) {
				t.Errorf("expected field %v, but got %v", field, returnedField)
			}
		})
	}
}

func TestColumnName(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	tests := []struct {
		fieldName    string
		expectedName string
	}{
		{
			fieldName:    "Age",
			expectedName: "age",
		},
		{
			fieldName:    "Name",
			expectedName: "name",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("field: %s", test.fieldName), func(t *testing.T) {
			field := fields.MustFind(test.fieldName)
			col := New(field)

			name, ok := col.Name()
			if !ok {
				t.Errorf("expected true, but got false")
			}
			if name != test.expectedName {
				t.Errorf("expected name %v, but got %v", test.fieldName, name)
			}
		})
	}
}

func TestColumnTable(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	tests := []struct {
		fieldName string
		tableName string
	}{
		{
			fieldName: "Age",
			tableName: "test_table",
		},
		{
			fieldName: "Name",
			tableName: "sample_table",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("field: %s", test.fieldName), func(t *testing.T) {
			field := fields.MustFind(test.fieldName)
			col := New(field, []Option{WithTable(test.tableName)}...)

			table, ok := col.Table()
			if !ok {
				t.Errorf("expected true, but got false")
			}
			if table != test.tableName {
				t.Errorf("expected table name %v, but got %v", test.tableName, table)
			}
		})
	}
}

func TestGenerateSQLQuery(t *testing.T) {
	tests := []struct {
		opt      *options
		expected string
	}{
		{
			opt:      &options{name: "columnName", table: "tableName"},
			expected: "tableName.columnName",
		},
		{
			opt:      &options{name: "columnName", table: ""},
			expected: "columnName",
		},
		{
			opt:      &options{name: "", table: "tableName"},
			expected: "tableName.",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("opt: %+v", test.opt), func(t *testing.T) {
			result := generateSQLQuery(test.opt)

			if result != test.expected {
				t.Errorf("expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGenerateToSQLFn(t *testing.T) {
	tests := []struct {
		sql      string
		alias    string
		expected string
	}{
		{
			sql:      "SELECT * FROM table",
			alias:    "",
			expected: "SELECT * FROM table",
		},
		{
			sql:      "SELECT * FROM table",
			alias:    "t",
			expected: "SELECT * FROM table AS t",
		},
		{
			sql:      "SELECT column FROM table",
			alias:    "",
			expected: "SELECT column FROM table",
		},
		{
			sql:      "SELECT column FROM table",
			alias:    "col",
			expected: "SELECT column FROM table AS col",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("sql: %q, alias: %q", test.sql, test.alias), func(t *testing.T) {
			fn := generateToSQLFn(test.sql, test.alias)
			result := fn(context.Background())
			if result != test.expected {
				t.Errorf("expected %v, but got %v", test.expected, result)
			}
		})
	}
}
