package api

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/column"
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type test struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt *time.Time
	Name      string
	PtrName   *string
	Age       int
	Bool      bool
	DeletedAt *time.Time
}

type testDto struct {
	ID        int        `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	Name      string     `json:"name"`
	PtrName   *string    `json:"ptr_name"`
	Age       int        `json:"age"`
	Bool      bool       `json:"bool"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// Zero allocates an input structure such that all pointer fields
// are fully allocated, i.e. rather than having a nil value,
// the pointer contains a pointer to an initialized value,
// e.g. an *int field will be a pointer to 0 instead of a nil pointer.
//
// zero does not allocate private fields.
func zero(obj interface{}) error {
	indirectVal := reflect.Indirect(reflect.ValueOf(obj))

	if !indirectVal.CanSet() {
		return fmt.Errorf("Input interface is not addressable (can't Set the memory address): %#v",
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

// MustZero will panic instead of return error
func mustZero(obj interface{}) {
	err := zero(obj)
	if err != nil {
		panic(err)
	}
}

func getModelAndFields[TModel any]() (*TModel, fmap.Storage, error) {
	model := new(TModel)
	mustZero(model)
	fields, err := fmap.GetFrom(model)
	if err != nil {
		return nil, nil, err
	}
	return model, fields, nil
}

func TestName(t *testing.T) {
	model, fields, err := getModelAndFields[test]()
	if err != nil {
		return
	}
	stor := types.NewEmptyColumnsStorage(fields)
	stor.Add(column.New(fields.MustFind("ID"), column.WithTable("test")))
	stor.Add(column.New(fields.MustFind("Name"), column.WithTable("test")))
	stor.Add(column.New(fields.MustFind("Age"), column.WithTable("test")))
	stor.Add(column.New(fields.MustFind("PtrName"), column.WithTable("test")))

	c, err := NewAPICore[test, testDto](stor)
	if err != nil {
		t.Fatal(err)
	}
	sorts := c.GetAvailableSorts()
	filters := c.GetAvailableFilters()
	_, _ = sorts, filters
	err = c.ValidateFilters("id:in:1,2,3,4,5,6||{id:eq:8||id:eq:9}$$ptr_name:ct:test")
	if err != nil {
		t.Fatal(err)
	}
	err = c.ValidateSorts("id-,age+,age1")
	if err != nil {
		t.Fatal(err)
	}
	apl := c.NewApplier()

	whereBuilder := linq.NewWhereBuilder(linq.NewCoreBuilder(model, stor))

	apl.ApplyFilters("id:in:1,2,3,4,5,6||{id:eq:8||id:eq:9}$$ptr_name:ct:test", whereBuilder)
	builder := sql.NewStringBuilder(context.Background(), "test", stor)
	whereBuilder.Apply(builder.WhereBuilder())
	_ = err
	fmt.Print(err)
}
