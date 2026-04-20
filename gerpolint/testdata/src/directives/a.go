package directives

import "github.com/insei/gerpo/types"

type Model struct {
	Age  int
	Name string
}

type whereTarget struct{}

func (whereTarget) Field(_ any) types.WhereOperation            { return nil }
func (whereTarget) Group(_ func(types.WhereTarget)) types.ANDOR { return nil }

func h() types.WhereTarget { return whereTarget{} }

func disableLine() {
	m := &Model{}

	h().Field(&m.Age).EQ("bad") //gerpolint:disable-line

	//gerpolint:disable-next-line
	h().Field(&m.Age).EQ("bad")
}

//gerpolint:disable
func inDisabledBlock() {
	m := &Model{}
	h().Field(&m.Age).EQ("bad")  // blanket disable
	h().Field(&m.Name).EQ(42)    // also silenced
}

//gerpolint:enable

func targetedDisable() {
	m := &Model{}

	// Only GPL001 disabled; would also be GPL003 (Contains on non-string) if
	// it weren't disabled — here we suppress the GPL001 that EQ would emit
	// and keep non-existent GPL003 path clean. Tests the targeted syntax.
	//gerpolint:disable-next-line=GPL001
	h().Field(&m.Age).EQ("bad")

	// Contains on non-string field produces GPL003; we only disable GPL001,
	// so GPL003 should still fire.
	//gerpolint:disable-next-line=GPL001
	h().Field(&m.Age).Contains("x") // want `GPL003: Contains is only applicable to string/\*string fields, got int`
}

func unknownRuleID() {
	m := &Model{}

	//gerpolint:disable-next-line=GPL999 // want `GPL-DIRECTIVE-UNKNOWN: unknown rule id in directive: GPL999`
	h().Field(&m.Age).EQ("bad") // want `GPL001: EQ: argument type string is not compatible with field type int`
}
