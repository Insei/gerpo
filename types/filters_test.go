package types

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/insei/fmap/v3"
)

func TestNewFilterManagerForField(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("TestField")

	testCases := []struct {
		name      string
		mockField fmap.Field
		expType   string
		expField  fmap.Field
	}{
		{
			name:      "create filter manager for field",
			mockField: field,
			expType:   "*column",
			expField:  field,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterManager := NewFilterManagerForField(tc.mockField)

			if _, ok := filterManager.(*column); !ok {
				t.Errorf("expected type %s, but got %T", tc.expType, filterManager)
			}

			col := filterManager.(*column)
			if !reflect.DeepEqual(col.field, tc.expField) {
				t.Errorf("expected field to be %v, but got %v", tc.expField, col.field)
			}
		})
	}
}

func TestAddFilterFn(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("TestField")

	testCases := []struct {
		name          string
		mockField     fmap.Field
		operation     Operation
		sqlGenFn      func(ctx context.Context, value any) (string, bool)
		testValue     any
		expSQL        string
		expNeedAppend bool
		expErr        bool
	}{
		{
			name:      "invalid filter function",
			mockField: field,
			operation: Operation("validOperation"),
			sqlGenFn: func(ctx context.Context, value any) (string, bool) {
				return fmt.Sprintf("SQL for %v", value), true
			},
			testValue:     "",
			expSQL:        "SQL for ",
			expNeedAppend: true,
			expErr:        true,
		},
		{
			name:      "valid filter function wrong type",
			mockField: field,
			operation: Operation("invalidOperation"),
			sqlGenFn: func(ctx context.Context, value any) (string, bool) {
				return "", false
			},
			testValue:     123,
			expSQL:        "",
			expNeedAppend: false,
			expErr:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterManager := NewFilterManagerForField(tc.mockField)

			filterManager.AddFilterFn(tc.operation, tc.sqlGenFn)

			filterFn, ok := filterManager.GetFilterFn(tc.operation)
			if !ok || filterFn == nil {
				t.Errorf("expected filter function for operation %v to exist", tc.operation)
				return
			}

			sql, needAppendValues, err := filterFn(context.Background(), tc.testValue)
			if (err != nil) != tc.expErr {
				t.Errorf("unexpected error result: got %v, expected error: %v", err != nil, tc.expErr)
			}

			if !tc.expErr {
				if sql != tc.expSQL || needAppendValues != tc.expNeedAppend {
					t.Errorf("unexpected result from filter function: sql=%v, needAppendValues=%v, err=%v", sql, needAppendValues, err)
				}
			}
		})
	}
}

func TestGetFilterFn(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("TestField")

	testCases := []struct {
		name         string
		operation    Operation
		sqlGenFn     func(ctx context.Context, value any) (string, bool)
		expAvailable bool
	}{
		{
			name:      "existing operation",
			operation: Operation("existingOperation"),
			sqlGenFn: func(ctx context.Context, value any) (string, bool) {
				return fmt.Sprintf("SQL for %v", value), true
			},
			expAvailable: true,
		},
		{
			name:         "non-existing operation",
			operation:    Operation("nonExistingOperation"),
			sqlGenFn:     nil,
			expAvailable: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterManager := NewFilterManagerForField(field)

			if tc.sqlGenFn != nil {
				filterManager.AddFilterFn(tc.operation, tc.sqlGenFn)
			}

			filterFn, available := filterManager.GetFilterFn(tc.operation)

			if available != tc.expAvailable {
				t.Errorf("unexpected availability result: got %v, expected %v", available, tc.expAvailable)
			}

			if available && filterFn == nil {
				t.Error("expected a valid filter function, but got nil")
			}

			if !available && filterFn != nil {
				t.Error("expected nil filter function for non-existing operation, but got a valid function")
			}
		})
	}
}

func TestGetAvailableFilterOperations(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("TestField")

	testCases := []struct {
		name          string
		operations    []Operation
		sqlGenFns     []func(ctx context.Context, value any) (string, bool)
		expOperations []Operation
	}{
		{
			name:          "no operations",
			operations:    []Operation{},
			sqlGenFns:     []func(ctx context.Context, value any) (string, bool){},
			expOperations: []Operation{},
		},
		{
			name: "single operation",
			operations: []Operation{
				Operation("singleOperation"),
			},
			sqlGenFns: []func(ctx context.Context, value any) (string, bool){
				func(ctx context.Context, value any) (string, bool) {
					return fmt.Sprintf("SQL for %v", value), true
				},
			},
			expOperations: []Operation{
				Operation("singleOperation"),
			},
		},
		{
			name: "multiple operations",
			operations: []Operation{
				Operation("firstOperation"),
				Operation("secondOperation"),
			},
			sqlGenFns: []func(ctx context.Context, value any) (string, bool){
				func(ctx context.Context, value any) (string, bool) {
					return fmt.Sprintf("SQL for %v", value), true
				},
				func(ctx context.Context, value any) (string, bool) {
					return fmt.Sprintf("SQL for %v", value), true
				},
			},
			expOperations: []Operation{
				Operation("firstOperation"),
				Operation("secondOperation"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterManager := NewFilterManagerForField(field)

			for i, operation := range tc.operations {
				filterManager.AddFilterFn(operation, tc.sqlGenFns[i])
			}

			availOperations := filterManager.GetAvailableFilterOperations()

			if len(availOperations) != len(tc.expOperations) {
				t.Errorf("unexpected number of available operations: got %d, expected %d", len(availOperations), len(tc.expOperations))
			}

			// Compare the actual content
			for _, op := range tc.expOperations {
				if !slices.Contains(availOperations, op) {
					t.Errorf("operation %v is expected but not found", op)
				}
			}
		})
	}
}

func TestIsAvailableFilterOperation(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("TestField")

	testCases := []struct {
		name           string
		operations     []Operation
		checkOperation Operation
		expAvailable   bool
	}{
		{
			name:           "operation not available",
			operations:     []Operation{},
			checkOperation: Operation("nonExistingOperation"),
			expAvailable:   false,
		},
		{
			name: "single operation available",
			operations: []Operation{
				Operation("existingOperation"),
			},
			checkOperation: Operation("existingOperation"),
			expAvailable:   true,
		},
		{
			name: "multiple operations, one available",
			operations: []Operation{
				Operation("existingOperation"),
				Operation("anotherOperation"),
			},
			checkOperation: Operation("existingOperation"),
			expAvailable:   true,
		},
		{
			name: "multiple operations, not available",
			operations: []Operation{
				Operation("firstOperation"),
				Operation("secondOperation"),
			},
			checkOperation: Operation("nonExistingOperation"),
			expAvailable:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterManager := NewFilterManagerForField(field)

			for _, operation := range tc.operations {
				filterManager.AddFilterFn(operation, func(ctx context.Context, value any) (string, bool) {
					return fmt.Sprintf("SQL for %v", value), true
				})
			}

			isAvailable := filterManager.IsAvailableFilterOperation(tc.checkOperation)

			if isAvailable != tc.expAvailable {
				t.Errorf("unexpected availability result for operation %v: got %v, expected %v", tc.checkOperation, isAvailable, tc.expAvailable)
			}
		})
	}
}
