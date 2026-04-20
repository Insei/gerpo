package unresolved_skip

import "github.com/insei/gerpo/types"

type Model struct {
	Age int
}

type whereTarget struct{}

func (whereTarget) Field(_ any) types.WhereOperation            { return nil }
func (whereTarget) Group(_ func(types.WhereTarget)) types.ANDOR { return nil }

func h() types.WhereTarget { return whereTarget{} }

func usesPtrVar() {
	m := &Model{}
	p := &m.Age
	// Default mode (skip): no diagnostic even though arg is a string.
	h().Field(p).EQ("bad")
}
