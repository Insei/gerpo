package column

import (
	"testing"

	"github.com/insei/fmap/v3"
	"github.com/stretchr/testify/assert"
)

type TestModel struct {
	Age  int
	Name string
}

func TestBuilderWithAlias(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")

	builder := &Builder{
		field: field,
	}
	builder2 := builder.WithAlias("testAlias")
	assert.Equal(t, builder, builder2)
}

func TestBuilderWithTable(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")

	builder := &Builder{
		field: field,
	}
	builder2 := builder.WithTable("testTable")
	assert.Equal(t, builder, builder2)
}

func TestBuilderWithColumnName(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")

	builder := &Builder{
		field: field,
	}
	builder2 := builder.WithColumnName("testColumnName")
	assert.Equal(t, builder, builder2)
}

func TestBuilderWithInsertProtection(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")

	builder := &Builder{
		field: field,
	}
	builder2 := builder.WithInsertProtection()
	assert.Equal(t, builder, builder2)
}

func TestBuilderWithUpdateProtection(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")

	builder := &Builder{
		field: field,
	}
	builder2 := builder.WithUpdateProtection()
	assert.Equal(t, builder, builder2)
}

func TestBuilderBuild(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")

	builder := &Builder{
		field: field,
	}
	columns := builder.Build()
	assert.NotEmpty(t, columns)
}
