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

	builder = builder.WithAlias("testAlias")
	if len(builder.opts) == 0 {
		t.Errorf("Expected opts to be set")
	}
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

	builder = builder.WithColumnName("testColumnName")
	if len(builder.opts) == 0 {
		t.Errorf("Expected opts to be set")
	}
}

func TestBuilderWithInsertProtection(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")

	builder := &Builder{
		field: field,
	}

	builder = builder.WithInsertProtection()
	if len(builder.opts) == 0 {
		t.Errorf("Expected opts to be set")
	}
}

func TestBuilderWithUpdateProtection(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Age")

	builder := &Builder{
		field: field,
	}

	builder = builder.WithUpdateProtection()
	if len(builder.opts) == 0 {
		t.Errorf("Expected opts to be set")
	}
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
