package column

import (
	"context"
	"fmt"
	"testing"

	"github.com/insei/fmap/v3"
)

func TestNew(t *testing.T) {
	type Test struct {
		Age     int
		TestAGE string
	}
	fields, _ := fmap.Get[Test]()

	cl := NewBuilder(fields.MustFind("Age")).
		WithTable("test").
		WithAlias("TestAge").
		Build()
	cl1 := NewBuilder(fields.MustFind("TestAGE")).
		WithTable("test").
		WithColumnName("super_test").
		WithAlias("test_age").Build()

	fmt.Print(cl.ToSQL(context.Background()), ", ", cl1.ToSQL(context.Background()))
}
