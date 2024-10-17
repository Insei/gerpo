package types

import (
	"context"
	"reflect"
	"testing"

	"github.com/insei/fmap/v3"
)

type TestColumn struct {
	*ColumnBase
}

func (t *TestColumn) Table() (string, bool) {
	return "", true
}

func (t *TestColumn) GetFilterFn(operation Operation) (func(ctx context.Context, value any) (string, bool, error), bool) {
	return t.ColumnBase.Filters.GetFilterFn(operation)
}

func (t *TestColumn) GetAvailableFilterOperations() []Operation {
	return t.ColumnBase.Filters.GetAvailableFilterOperations()
}

func (t *TestColumn) IsAvailableFilterOperation(operation Operation) bool {
	return t.ColumnBase.Filters.IsAvailableFilterOperation(operation)
}

func (t *TestColumn) GetAllowedActions() []SQLAction {
	return t.ColumnBase.AllowedActions
}

func (t *TestColumn) ToSQL(ctx context.Context) string {
	return t.ColumnBase.ToSQL(ctx)
}

func (t *TestColumn) GetPtr(model any) any {
	return t.ColumnBase.GetPtr(model)
}

func (t *TestColumn) IsAllowedAction(action SQLAction) bool {
	return t.ColumnBase.IsAllowedAction(action)
}

func (t *TestColumn) GetField() fmap.Field {
	return t.ColumnBase.Field
}

func (t *TestColumn) Name() (string, bool) {
	return t.GetField().GetName(), true
}

type TestModel struct {
	TestField    int
	AnotherField string
}

func TestColumnBaseIsAllowedAction(t *testing.T) {
	type Test struct {
		Age     int
		TestAGE string
	}
	fields, _ := fmap.Get[Test]()

	field := fields.MustFind("Age")
	toSQLFn := func(ctx context.Context) string {
		return "test_sql"
	}
	cb := NewColumnBase(field, toSQLFn, nil)

	cases := []struct {
		name          string
		actions       []SQLAction
		action        SQLAction
		expectedError bool
	}{
		{
			name:          "check allowed action",
			actions:       []SQLAction{SQLActionSelect},
			action:        SQLActionSelect,
			expectedError: false,
		},
		{
			name:          "check disallowed action",
			actions:       []SQLAction{SQLActionInsert},
			action:        SQLActionSelect,
			expectedError: true,
		},
		{
			name:          "check multiple allowed actions",
			actions:       []SQLAction{SQLActionSelect, SQLActionInsert},
			action:        SQLActionSelect,
			expectedError: false,
		},
		{
			name:          "check empty allowed actions",
			actions:       []SQLAction{},
			action:        SQLActionSelect,
			expectedError: true,
		},
		{
			name:          "check nil allowed actions",
			actions:       nil,
			action:        SQLActionSelect,
			expectedError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cb.AllowedActions = tc.actions
			if tc.expectedError {
				if cb.IsAllowedAction(tc.action) {
					t.Errorf("expected IsAllowedAction to return false for disallowed action, but got true")
				}
			} else {
				if !cb.IsAllowedAction(tc.action) {
					t.Errorf("expected IsAllowedAction to return true for allowed action, but got false")
				}
			}
		})
	}
}

func TestColumnsStorageGetByFieldPtr(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	cs := NewEmptyColumnsStorage(fields)
	testModel := &TestModel{
		TestField: 1,
	}

	cases := []struct {
		name          string
		field         string
		column        Column
		model         *TestModel
		fieldPtr      any
		expectedError bool
	}{
		{
			name:          "check existing field",
			field:         "TestField",
			column:        &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("TestField"), nil, nil)},
			model:         testModel,
			fieldPtr:      &testModel.TestField,
			expectedError: false,
		},
		{
			name:          "check non-existing field",
			field:         "NonExistingField",
			column:        nil,
			model:         &TestModel{},
			fieldPtr:      &TestModel{},
			expectedError: true,
		},
		{
			name:          "check nil model",
			field:         "TestField",
			column:        &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("TestField"), nil, nil)},
			model:         nil,
			fieldPtr:      &TestModel{},
			expectedError: true,
		},
		{
			name:          "check field pointer of wrong type",
			field:         "TestField",
			column:        &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("TestField"), nil, nil)},
			model:         &TestModel{},
			fieldPtr:      new(string),
			expectedError: true,
		},
		{
			name:          "check field pointer of wrong type",
			field:         "TestField",
			column:        nil,
			model:         &TestModel{},
			fieldPtr:      new(string),
			expectedError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.column != nil {
				cs.Add(fields.MustFind(tc.field), tc.column)
			}
			gotColumn, err := cs.GetByFieldPtr(tc.model, tc.fieldPtr)
			if tc.expectedError {
				if err == nil {
					t.Errorf("expected GetByFieldPtr to return an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected GetByFieldPtr to return no error, but got: %v", err)
				}
				if gotColumn != tc.column {
					t.Errorf("expected GetByFieldPtr to return the correct column, but got: %v", gotColumn)
				}
			}
		})
	}
}

func TestColumnsStorageGet(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	cs := NewEmptyColumnsStorage(fields)
	column1 := &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("TestField"), nil, nil)}
	column2 := &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("AnotherField"), nil, nil)}

	cases := []struct {
		name           string
		field          fmap.Field
		column         Column
		expectedFound  bool
		expectedColumn Column
	}{
		{
			name:           "existing field",
			field:          fields.MustFind("TestField"),
			column:         column1,
			expectedFound:  true,
			expectedColumn: column1,
		},
		{
			name:           "non-existing field",
			field:          fields.MustFind("NonExistingField"),
			column:         nil,
			expectedFound:  false,
			expectedColumn: nil,
		},
		{
			name:           "another existing field",
			field:          fields.MustFind("AnotherField"),
			column:         column2,
			expectedFound:  true,
			expectedColumn: column2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.column != nil {
				cs.Add(tc.field, tc.column)
			}
			gotColumn, ok := cs.Get(tc.field)
			if ok != tc.expectedFound {
				t.Errorf("expected Get to return %v for field %q, but got %v", tc.expectedFound, tc.field.GetName(), ok)
			}
			if gotColumn != tc.expectedColumn {
				t.Errorf("expected Get to return column %v for field %q, but got %v", tc.expectedColumn, tc.field.GetName(), gotColumn)
			}
		})
	}
}

func TestColumnsStorageAdd(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	cs := NewEmptyColumnsStorage(fields)

	column1 := &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("TestField"), nil, nil)}
	column2 := &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("AnotherField"), nil, nil)}
	column3 := &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("TestField"), nil, nil)}
	column3.AllowedActions = []SQLAction{SQLActionInsert, SQLActionSelect, SQLActionUpdate, SQLActionGroup, SQLActionSort}

	cases := []struct {
		name           string
		field          fmap.Field
		column         Column
		expectedLenS   int
		expectedLenAct int
		expectedAct    map[SQLAction][]Column
	}{
		{
			name:           "add first column",
			field:          fields.MustFind("TestField"),
			column:         column1,
			expectedLenS:   1,
			expectedLenAct: 0,
			expectedAct:    map[SQLAction][]Column{},
		},
		{
			name:           "add column with one allowed action",
			field:          fields.MustFind("AnotherField"),
			column:         column2,
			expectedLenS:   2,
			expectedLenAct: 0,
			expectedAct:    map[SQLAction][]Column{},
		},
		{
			name:           "add column with all allowed actions",
			field:          fields.MustFind("TestField"),
			column:         column3,
			expectedLenS:   3,
			expectedLenAct: 5,
			expectedAct: map[SQLAction][]Column{
				SQLActionInsert: {column3},
				SQLActionSelect: {column3},
				SQLActionUpdate: {column3},
				SQLActionGroup:  {column3},
				SQLActionSort:   {column3},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cs.Add(tc.field, tc.column)
			if len(cs.s) != tc.expectedLenS {
				t.Errorf("expected Add to add %d columns to the storage, but got %d", tc.expectedLenS, len(cs.s))
			}
			if len(cs.act) != tc.expectedLenAct {
				t.Errorf("expected Add to add %d actions, but got %d", tc.expectedLenAct, len(cs.act))
			}
			for action, expectedColumns := range tc.expectedAct {
				gotColumns, ok := cs.act[action]
				if !ok {
					t.Errorf("expected Add to add action %q, but it was not found", action)
				}
				if !reflect.DeepEqual(gotColumns, expectedColumns) {
					t.Errorf("expected Add to add columns %v for action %q, but got %v", expectedColumns, action, gotColumns)
				}
			}
		})
	}
}

func TestColumnsStorageAsSlice(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	cs := NewEmptyColumnsStorage(fields)

	column1 := &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("TestField"), nil, nil)}
	column2 := &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("AnotherField"), nil, nil)}
	column3 := &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("TestField"), nil, nil)}

	cases := []struct {
		name            string
		columns         []Column
		expectedLength  int
		expectedColumns []Column
	}{
		{
			name:            "empty storage",
			columns:         nil,
			expectedLength:  0,
			expectedColumns: nil,
		},
		{
			name:            "one column",
			columns:         []Column{column1},
			expectedLength:  1,
			expectedColumns: []Column{column1},
		},
		{
			name:            "multiple columns",
			columns:         []Column{column2, column3},
			expectedLength:  3,
			expectedColumns: []Column{column1, column2, column3},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, column := range tc.columns {
				cs.Add(column.GetField(), column)
			}
			gotColumns := cs.AsSlice()
			if len(gotColumns) != tc.expectedLength {
				t.Errorf("expected AsSlice to return %d columns, but got %d", tc.expectedLength, len(gotColumns))
			}
			for i, gotColumn := range gotColumns {
				if gotColumn != tc.expectedColumns[i] {
					t.Errorf("expected AsSlice to return column %v at index %d, but got %v", tc.expectedColumns[i], i, gotColumn)
				}
			}
		})
	}
}

func TestColumnsStorageAsSliceByAction(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	cs := NewEmptyColumnsStorage(fields)

	column1 := &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("TestField"), nil, nil)}
	column1.AllowedActions = []SQLAction{SQLActionInsert, SQLActionSelect}

	column2 := &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("AnotherField"), nil, nil)}
	column2.AllowedActions = []SQLAction{SQLActionInsert, SQLActionUpdate}

	column3 := &TestColumn{ColumnBase: NewColumnBase(fields.MustFind("TestField"), nil, nil)}
	column3.AllowedActions = []SQLAction{SQLActionSelect}

	cases := []struct {
		name            string
		columns         []Column
		action          SQLAction
		expectedLength  int
		expectedColumns []Column
	}{
		{
			name:            "no columns for action",
			columns:         nil,
			action:          SQLActionInsert,
			expectedLength:  0,
			expectedColumns: nil,
		},
		{
			name:            "one column for action",
			columns:         []Column{column1},
			action:          SQLActionInsert,
			expectedLength:  1,
			expectedColumns: []Column{column1},
		},
		{
			name:            "multiple columns for action",
			columns:         []Column{column2},
			action:          SQLActionInsert,
			expectedLength:  2,
			expectedColumns: []Column{column1, column2},
		},
		{
			name:            "columns for different actions",
			columns:         []Column{column3},
			action:          SQLActionSelect,
			expectedLength:  2,
			expectedColumns: []Column{column1, column3},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, column := range tc.columns {
				cs.Add(column.GetField(), column)
			}
			gotColumns := cs.AsSliceByAction(tc.action)
			if len(gotColumns) != tc.expectedLength {
				t.Errorf("expected AsSliceByAction to return %d columns for action %v, but got %d", tc.expectedLength, tc.action, len(gotColumns))
			}
			for i, gotColumn := range gotColumns {
				if gotColumn != tc.expectedColumns[i] {
					t.Errorf("expected AsSliceByAction to return column %v at index %d for action %v, but got %v", tc.expectedColumns[i], i, tc.action, gotColumn)
				}
			}
		})
	}
}
