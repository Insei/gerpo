package gerpo

import (
	"fmt"
	"reflect"

	"github.com/insei/fmap/v3"
)

// Zero allocates an input structure such that all pointer fields
// are fully allocated, i.e. rather than having a nil value,
// the pointer contains a pointer to an initialized value,
// e.g. an *int field will be a pointer to 0 instead of a nil pointer.
//
// zero does not allocate private fields.
func zero(obj interface{}) error {
	indirectVal := reflect.Indirect(reflect.ValueOf(obj))

	if !indirectVal.CanSet() {
		return fmt.Errorf("input interface is not addressable (can't Set the memory address): %#v",
			obj)
	}
	if indirectVal.Kind() != reflect.Struct {
		return fmt.Errorf("allocate.Zero currently only works with [pointers to] structs, not type %v",
			indirectVal.Kind())
	}

	// allocate each of the structs fields
	var err error
	for i := 0; i < indirectVal.NumField(); i++ {
		field := indirectVal.Field(i)

		// pre-allocate pointer fields
		if field.Kind() == reflect.Ptr && field.IsNil() {
			if field.CanSet() {
				field.Set(reflect.New(field.Type().Elem()))
			}
		}

		indirectField := reflect.Indirect(field)
		switch indirectField.Kind() {
		case reflect.Map:
			indirectField.Set(reflect.MakeMap(indirectField.Type()))
		case reflect.Struct:
			// recursively allocate each of the structs embedded fields
			if field.Kind() == reflect.Ptr {
				err = zero(field.Interface())
			} else {
				// field of Struct can always use field.Addr()
				fieldAddr := field.Addr()
				if fieldAddr.CanInterface() {
					err = zero(fieldAddr.Interface())
				} else {
					err = fmt.Errorf("struct field can't interface, %#v", fieldAddr)
				}
			}
		}
		if err != nil {
			return err
		}
	}
	return err
}

func getModelAndFields[TModel any]() (*TModel, fmap.Storage, error) {
	model := new(TModel)
	err := zero(model)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to zero model: %w", err)
	}
	fields, err := fmap.GetFrom(model)
	if err != nil {
		return nil, nil, err
	}
	return model, fields, nil
}
