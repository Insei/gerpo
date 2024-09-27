package virtual

import (
	"context"
	"fmt"
	"testing"

	"github.com/insei/fmap/v3"
)

type Test struct {
	Age     int
	TestAGE *bool
}

func TestBuilder(t *testing.T) {
	fields, _ := fmap.Get[Test]()
	field := fields.MustFind("TestAGE")
	b := NewBuilder(field).
		WithSQL(func(ctx context.Context) string {
			return "courses.created_at > now() AS is_new"
		}).
		WithBoolEqFilter("courses.created_at > now()", "courses.created_at < now()", "")
	fmt.Println(b)
}
