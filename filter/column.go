package filter

import (
	"fmt"
	"reflect"

	"github.com/insei/fmap/v3"
)

type SQLFilterManager interface {
	AddFilterFn(operation Operation, sqlGenFn func(value any) (string, bool))
	SQLFilterGetter
}

type SQLFilterGetter interface {
	GetFilterFn(operation Operation) (func(value any) (string, bool, error), bool)
}

type column struct {
	field      fmap.Field
	operations map[Operation]func(value any) (string, bool, error)
}

func NewForField(field fmap.Field) SQLFilterManager {
	return &column{
		operations: map[Operation]func(value any) (string, bool, error){},
		field:      field,
	}
}

func (c *column) AddFilterFn(operation Operation, sqlGenFn func(value any) (string, bool)) {
	c.operations[operation] = func(value any) (string, bool, error) {
		vType := reflect.TypeOf(value)
		for vType.Kind() == reflect.Ptr {
			vType = vType.Elem()
		}
		if c.field.GetDereferencedType() != vType {
			return "", false, fmt.Errorf("filter value type not a valid for field")
		}
		sql, needAppendValues := sqlGenFn(value)
		return sql, needAppendValues, nil
	}
}

func (c *column) GetFilterFn(operation Operation) (func(value any) (string, bool, error), bool) {
	if opFn, ok := c.operations[operation]; ok && opFn != nil {
		return opFn, true
	}
	return nil, false
}
