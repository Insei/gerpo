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

func mutate(p *[]any)        { *p = append(*p, 1) }
func getSlice() []any        { return []any{1, 2} }

func positives() {
	m := &Model{}

	h().Field(&m.Age).In(1, 2, 3)
	h().Field(&m.Age).NotIn()
	h().Field(&m.Name).In("a", "b")

	// *string field accepts string, *string, nil.
	s := "x"
	h().Field(&m.Email).In("a", "b", nil)
	h().Field(&m.Email).In(&s)

	// Inline []any{...}... — elements are type-checked individually.
	h().Field(&m.Age).In([]any{1, 2, 3}...)
	h().Field(&m.Name).In([]any{"a", "b"}...)

	// []any through a single-assignment local var — elements recovered.
	ages := []any{10, 20, 30}
	h().Field(&m.Age).In(ages...)

	// append-chain accumulator from `var t []any`.
	var t []any
	t = append(t, 1)
	t = append(t, 2, 3)
	h().Field(&m.Age).In(t...)

	// Short-declared accumulator + sequential appends.
	u := []any{}
	u = append(u, "a")
	u = append(u, "b", "c")
	h().Field(&m.Name).In(u...)

	// append(v, literal...) with an inline spread — elements expand.
	var w []any
	w = append(w, 1)
	w = append(w, []any{2, 3}...)
	h().Field(&m.Age).In(w...)

	// Auto-unwrap: gerpo accepts a single slice argument without `...`.
	h().Field(&m.Age).In([]any{1, 2, 3})
	h().Field(&m.Name).In([]any{"a", "b"})

	idsNoSpread := []any{10, 20}
	h().Field(&m.Age).In(idsNoSpread)

	var p []any
	p = append(p, 1, 2)
	h().Field(&m.Age).In(p)
}

func negatives() {
	m := &Model{}

	h().Field(&m.Age).In(1, "2", 3)    // want `GPL002: In: argument type string is not compatible with field type int`
	h().Field(&m.Age).NotIn(1, 2, nil) // want `GPL002: NotIn: argument type untyped nil is not compatible with field type int`
	h().Field(&m.Name).In(1, "b")      // want `GPL002: In: argument type int is not compatible with field type string`
	h().Field(&m.Email).In(1)          // want `GPL002: In: argument type int is not compatible with field type \*string`

	// Inline []any spread — each element checked; one bad entry flagged.
	h().Field(&m.Age).In([]any{1, "bad", 3}...) // want `GPL002: In: argument type string is not compatible with field type int`

	// []any through a single-assignment var — element type is statically resolvable.
	names := []any{"a", 42} // want `GPL002: In: argument type int is not compatible with field type string`
	h().Field(&m.Name).In(names...)

	// append-chain with a bad element mid-batch — flagged on the element itself.
	var t []any
	t = append(t, 1, "bad", 3) // want `GPL002: In: argument type string is not compatible with field type int`
	h().Field(&m.Age).In(t...)

	// Sequential appends — a bad element in a later append is still caught.
	var u []any
	u = append(u, 1)
	u = append(u, "bad") // want `GPL002: In: argument type string is not compatible with field type int`
	h().Field(&m.Age).In(u...)

	// No-spread form: bad element inside the slice is still caught.
	h().Field(&m.Age).In([]any{1, "nope", 3}) // want `GPL002: In: argument type string is not compatible with field type int`

	// No-spread with typed slice that mismatches the field.
	strs := []string{"a"}
	h().Field(&m.Age).In(strs) // want `GPL002: In: slice element type string is not compatible with field type int`
}

func blockedCases() {
	m := &Model{}

	// &v forbids static tracking — falls back to GPL005.
	var t []any
	mutate(&t)
	t = append(t, 1)
	h().Field(&m.Age).In(t...) // want `GPL005: In: slice element type is .any.; static check skipped`

	// Reassignment from a non-literal, non-append RHS blocks the variable.
	u := []any{1}
	u = getSlice()
	h().Field(&m.Age).In(u...) // want `GPL005: In: slice element type is .any.; static check skipped`

	// Index-write mutates elements we can't track — block.
	w := []any{1, 2}
	w[1] = "x"
	h().Field(&m.Age).In(w...) // want `GPL005: In: slice element type is .any.; static check skipped`
}
