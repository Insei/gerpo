package types

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/insei/fmap/v3"
)

type column struct {
	field      fmap.Field
	operations map[Operation]func(ctx context.Context, value any) (string, bool, error)
	avail      []Operation
}

// NewFilterManagerForField creates a new SQLFilterManager for the given field, managing filter functions and operations.
func NewFilterManagerForField(field fmap.Field) SQLFilterManager {
	return &column{
		operations: map[Operation]func(ctx context.Context, value any) (string, bool, error){},
		field:      field,
	}
}

func (c *column) AddFilterFn(operation Operation, sqlGenFn func(ctx context.Context, value any) (string, bool)) {
	c.avail = append(c.avail, operation)
	c.operations[operation] = func(ctx context.Context, value any) (string, bool, error) {
		if c.field.GetType().Kind() == reflect.Ptr && value == nil &&
			(operation == OperationEQ || operation == OperationNEQ) {
			sql, needAppendValues := sqlGenFn(ctx, value)
			return sql, needAppendValues, nil
		}

		vType := reflect.TypeOf(value)
		for vType.Kind() == reflect.Ptr {
			vType = vType.Elem()
		}

		if vType.Kind() == reflect.Slice {
			valuesOf := reflect.ValueOf(value)
			if valuesOf.Len() < 1 {
				return "", false, nil
			}
			arrVal := valuesOf.Index(0).Interface()
			vType = reflect.TypeOf(arrVal)
			if arrValTypeOf := reflect.TypeOf(arrVal); arrValTypeOf.Kind() == reflect.Slice {
				value = reflect.ValueOf(value).Index(0).Interface()
				vType = arrValTypeOf.Elem()
			}
		}
		if c.field.GetDereferencedType() != vType {
			return "", false, fmt.Errorf("whereSQL value[%s] type not a valid for field \"%s\" of type [%s]. Value: %v", vType.Name(), c.field.GetName(), c.field.GetType().Name(), value)
		}
		sql, needAppendValues := sqlGenFn(ctx, value)
		return sql, needAppendValues, nil
	}
}

func (c *column) GetFilterFn(operation Operation) (func(ctx context.Context, value any) (string, bool, error), bool) {
	if opFn, ok := c.operations[operation]; ok && opFn != nil {
		return opFn, true
	}
	return nil, false
}

func (c *column) GetAvailableFilterOperations() []Operation {
	return c.avail
}

func (c *column) IsAvailableFilterOperation(operation Operation) bool {
	return slices.Contains(c.avail, operation)
}
