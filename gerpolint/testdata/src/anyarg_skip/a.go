package anyarg_skip

import "github.com/insei/gerpo/types"

type Model struct {
	Age int
}

type whereTarget struct{}

func (whereTarget) Field(_ any) types.WhereOperation            { return nil }
func (whereTarget) Group(_ func(types.WhereTarget)) types.ANDOR { return nil }

func h() types.WhereTarget { return whereTarget{} }

func anyArgSkipped() {
	m := &Model{}
	var v any = 18
	// With -any-arg=skip, no diagnostic.
	h().Field(&m.Age).EQ(v)
}
