package unresolved

import "github.com/insei/gerpo/types"

type Model struct {
	Age int
}

type whereTarget struct{}

func (whereTarget) Field(_ any) types.WhereOperation            { return nil }
func (whereTarget) Group(_ func(types.WhereTarget)) types.ANDOR { return nil }

func h() types.WhereTarget { return whereTarget{} }

func getPtr(m *Model) any { return &m.Age }

func viaVariable() {
	m := &Model{}
	p := &m.Age
	h().Field(p).EQ("bad") // want `GPL004: cannot statically resolve field type for Field\(\.\.\.\)\.EQ\(\.\.\.\)`
}

func viaHelper() {
	m := &Model{}
	h().Field(getPtr(m)).EQ("bad") // want `GPL004: cannot statically resolve field type for Field\(\.\.\.\)\.EQ\(\.\.\.\)`
}
