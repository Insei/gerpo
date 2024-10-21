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

func NewFilterManagerForField(field fmap.Field) SQLFilterManager {
	return &column{
		operations: map[Operation]func(ctx context.Context, value any) (string, bool, error){},
		field:      field,
	}
}

func (c *column) AddFilterFn(operation Operation, sqlGenFn func(ctx context.Context, value any) (string, bool)) {
	c.avail = append(c.avail, operation)
	c.operations[operation] = func(ctx context.Context, value any) (string, bool, error) {
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
		}
		if c.field.GetDereferencedType() != vType {
			return "", false, fmt.Errorf("whereSQL value type not a valid for field")
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
