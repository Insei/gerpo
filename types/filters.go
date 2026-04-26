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
	operations map[Operation]func(ctx context.Context, value any) (string, []any, error)
	avail      []Operation
}

// NewFilterManagerForField creates a new SQLFilterManager for the given field, managing filter functions and operations.
func NewFilterManagerForField(field fmap.Field) SQLFilterManager {
	return &column{
		operations: map[Operation]func(ctx context.Context, value any) (string, []any, error){},
		field:      field,
	}
}

// AddFilterFn adapts the legacy (sql, bool) shape to the args-based one used internally.
// `bool == true` → the user value is appended as a single bound arg (slices are expanded
// by the consumer); `bool == false` → no value is appended.
func (c *column) AddFilterFn(operation Operation, sqlGenFn func(ctx context.Context, value any) (string, bool)) {
	c.AddFilterFnArgs(operation, func(ctx context.Context, value any) (string, []any, error) {
		sql, appendValue := sqlGenFn(ctx, value)
		if !appendValue {
			return sql, nil, nil
		}
		return sql, []any{value}, nil
	})
}

func (c *column) AddFilterFnArgs(operation Operation, sqlGenFn func(ctx context.Context, value any) (string, []any, error)) {
	c.avail = append(c.avail, operation)
	c.operations[operation] = func(ctx context.Context, value any) (string, []any, error) {
		if c.field.GetType().Kind() == reflect.Ptr && value == nil &&
			(operation == OperationEQ || operation == OperationNotEQ) {
			return sqlGenFn(ctx, value)
		}

		vType := reflect.TypeOf(value)
		for vType != nil && vType.Kind() == reflect.Ptr {
			vType = vType.Elem()
		}

		if vType != nil && vType.Kind() == reflect.Slice {
			valuesOf := reflect.ValueOf(value)
			if valuesOf.Len() < 1 {
				return "", nil, nil
			}
			arrVal := valuesOf.Index(0).Interface()
			vType = reflect.TypeOf(arrVal)
			if arrValTypeOf := reflect.TypeOf(arrVal); arrValTypeOf.Kind() == reflect.Slice {
				value = reflect.ValueOf(value).Index(0).Interface()
				vType = arrValTypeOf.Elem()
			}
		}
		if vType != nil && c.field.GetDereferencedType() != vType {
			return "", nil, fmt.Errorf("whereSQL value[%s] type not a valid for field \"%s\" of type [%s]. Value: %v", vType.Name(), c.field.GetName(), c.field.GetType().Name(), value)
		}
		return sqlGenFn(ctx, value)
	}
}

// AddFilterFnArgsRaw skips the runtime reflect.Type equality check that
// AddFilterFnArgs imposes. Used by the global filters.Registry path so custom
// types and string-aliases work — the registry already binds operators to the
// types that should accept them.
//
// Empty slices still collapse to (sql="", nil, nil) so the caller emits no SQL,
// matching the consumer's expectations.
func (c *column) AddFilterFnArgsRaw(operation Operation, sqlGenFn func(ctx context.Context, value any) (string, []any, error)) {
	c.avail = append(c.avail, operation)
	c.operations[operation] = func(ctx context.Context, value any) (string, []any, error) {
		// Same empty-slice short-circuit as AddFilterFnArgs — without it the
		// In/NotIn fragments would render placeholders for zero values.
		if vType := reflect.TypeOf(value); vType != nil && vType.Kind() == reflect.Slice {
			if reflect.ValueOf(value).Len() < 1 {
				return "", nil, nil
			}
		}
		return sqlGenFn(ctx, value)
	}
}

func (c *column) GetFilterFn(operation Operation) (func(ctx context.Context, value any) (string, []any, error), bool) {
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
