package filter

import (
	"context"
	"fmt"
	"reflect"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/types"
)

type column struct {
	field      fmap.Field
	operations map[types.Operation]func(ctx context.Context, value any) (string, bool, error)
}

func NewForField(field fmap.Field) types.SQLFilterManager {
	return &column{
		operations: map[types.Operation]func(ctx context.Context, value any) (string, bool, error){},
		field:      field,
	}
}

func (c *column) AddFilterFn(operation types.Operation, sqlGenFn func(ctx context.Context, value any) (string, bool)) {
	c.operations[operation] = func(ctx context.Context, value any) (string, bool, error) {
		vType := reflect.TypeOf(value)
		for vType.Kind() == reflect.Ptr {
			vType = vType.Elem()
		}
		if c.field.GetDereferencedType() != vType {
			return "", false, fmt.Errorf("filter value type not a valid for field")
		}
		sql, needAppendValues := sqlGenFn(ctx, value)
		return sql, needAppendValues, nil
	}
}

func (c *column) GetFilterFn(operation types.Operation) (func(ctx context.Context, value any) (string, bool, error), bool) {
	if opFn, ok := c.operations[operation]; ok && opFn != nil {
		return opFn, true
	}
	return nil, false
}
