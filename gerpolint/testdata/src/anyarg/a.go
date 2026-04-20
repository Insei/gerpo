package anyarg

import "github.com/insei/gerpo/types"

type Model struct {
	Age int
}

type whereTarget struct{}

func (whereTarget) Field(_ any) types.WhereOperation            { return nil }
func (whereTarget) Group(_ func(types.WhereTarget)) types.ANDOR { return nil }

func h() types.WhereTarget { return whereTarget{} }

func produce() any { return 18 }

func viaVar() {
	m := &Model{}
	var v any = 18
	h().Field(&m.Age).EQ(v) // want `GPL005: argument to EQ has static type .any.; static check skipped`
}

func viaCall() {
	m := &Model{}
	h().Field(&m.Age).EQ(produce()) // want `GPL005: argument to EQ has static type .any.; static check skipped`
}
