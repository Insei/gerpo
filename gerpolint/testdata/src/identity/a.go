package identity

// MyBuilder has the same EQ/In method names as gerpo's WhereOperation, but
// lives in a different package. The linter must not fire on these.
type MyBuilder struct{}

func (MyBuilder) Field(_ any) MyBuilder     { return MyBuilder{} }
func (MyBuilder) EQ(_ any) MyBuilder        { return MyBuilder{} }
func (MyBuilder) In(_ ...any) MyBuilder     { return MyBuilder{} }
func (MyBuilder) Contains(_ any) MyBuilder  { return MyBuilder{} }
func (MyBuilder) NotContains(_ any) MyBuilder { return MyBuilder{} }

func b() MyBuilder { return MyBuilder{} }

type Model struct {
	Age int
}

func checks() {
	m := &Model{}

	// Would fire if the linter mis-identified this as gerpo — arguments
	// deliberately violate the would-be rule.
	b().Field(&m.Age).EQ("not-int")
	b().Field(&m.Age).In(1, "2", 3)
	b().Field(&m.Age).Contains("anything")

	// Chain without Field(...) prefix — the linter requires Field as the
	// receiver, so even matching EQ names should be ignored.
	b().EQ("literally anything")
}
