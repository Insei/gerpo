package chains

import "github.com/insei/gerpo/types"

type Model struct {
	Age  int
	Name string
}

type whereTarget struct{}

func (whereTarget) Field(_ any) types.WhereOperation            { return nil }
func (whereTarget) Group(_ func(types.WhereTarget)) types.ANDOR { return &andor{} }

type andor struct{}

func (*andor) AND() types.WhereTarget { return whereTarget{} }
func (*andor) OR() types.WhereTarget  { return whereTarget{} }

func h() types.WhereTarget { return whereTarget{} }

func positives() {
	m := &Model{}

	h().Field(&m.Age).EQ(18)
	h().Field(&m.Age).GT(10)
	h().Field(&m.Name).EQ("a")

	h().Group(func(t types.WhereTarget) {
		t.Field(&m.Age).EQ(18)
		t.Group(func(t2 types.WhereTarget) {
			t2.Field(&m.Name).EQ("a")
		})
	})
}

func negatives() {
	m := &Model{}

	h().Field(&m.Age).EQ("bad")  // want `GPL001: EQ: argument type string is not compatible with field type int`
	h().Field(&m.Name).EQ(42)    // want `GPL001: EQ: argument type int is not compatible with field type string`

	h().Group(func(t types.WhereTarget) {
		t.Field(&m.Age).EQ("bad") // want `GPL001: EQ: argument type string is not compatible with field type int`
		t.Group(func(t2 types.WhereTarget) {
			t2.Field(&m.Age).EQ("bad") // want `GPL001: EQ: argument type string is not compatible with field type int`
		})
	})
}
