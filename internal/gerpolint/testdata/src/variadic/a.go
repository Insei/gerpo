package variadic

import "github.com/insei/gerpo/types"

type Model struct {
	Age   int
	Name  string
	Email *string
}

type whereTarget struct{}

func (whereTarget) Field(_ any) types.WhereOperation            { return nil }
func (whereTarget) Group(_ func(types.WhereTarget)) types.ANDOR { return nil }

func h() types.WhereTarget { return whereTarget{} }

func positives() {
	m := &Model{}

	h().Field(&m.Age).In(1, 2, 3)
	h().Field(&m.Age).NotIn()
	h().Field(&m.Name).In("a", "b")

	// *string field accepts string, *string, nil.
	s := "x"
	h().Field(&m.Email).In("a", "b", nil)
	h().Field(&m.Email).In(&s)
}

func negatives() {
	m := &Model{}

	h().Field(&m.Age).In(1, "2", 3)    // want `GPL002: In: argument type string is not compatible with field type int`
	h().Field(&m.Age).NotIn(1, 2, nil) // want `GPL002: NotIn: argument type untyped nil is not compatible with field type int`
	h().Field(&m.Name).In(1, "b")      // want `GPL002: In: argument type int is not compatible with field type string`
	h().Field(&m.Email).In(1)          // want `GPL002: In: argument type int is not compatible with field type \*string`
}
