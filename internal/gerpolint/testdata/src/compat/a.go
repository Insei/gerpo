package compat

import (
	"time"

	"github.com/insei/gerpo/types"
)

type UUID [16]byte

type Model struct {
	ID        UUID
	Age       int
	Age32     int32
	Age64     int64
	Ticket    uint32
	Price     float64
	Active    bool
	Name      string
	Email     *string
	CreatedAt time.Time
}

// whereTarget is a stub — the real gerpo builds it internally. The analyzer
// only cares that the method call chain goes through types.WhereTarget /
// types.WhereOperation, so any implementation works.
type whereTarget struct{}

func (whereTarget) Field(_ any) types.WhereOperation { return nil }
func (whereTarget) Group(_ func(types.WhereTarget)) types.ANDOR { return nil }

func h() types.WhereTarget { return whereTarget{} }

func positives() {
	m := &Model{}

	// Scalar × primitive: int
	h().Field(&m.Age).EQ(18)
	h().Field(&m.Age).NotEQ(0)
	h().Field(&m.Age).LT(65)
	h().Field(&m.Age).LTE(65)
	h().Field(&m.Age).GT(0)
	h().Field(&m.Age).GTE(18)

	// Fixed-width ints / unsigned
	h().Field(&m.Age32).EQ(int32(5))
	h().Field(&m.Age64).EQ(int64(5))
	h().Field(&m.Ticket).EQ(uint32(5))
	// Untyped int literal to sized int: representable → OK.
	h().Field(&m.Age32).EQ(5)
	h().Field(&m.Ticket).EQ(0)

	// UUID array literal (arrays compare by elements).
	var id UUID
	h().Field(&m.ID).EQ(id)

	// float64
	h().Field(&m.Price).EQ(3.14)
	h().Field(&m.Price).GT(0.0)

	// bool
	h().Field(&m.Active).EQ(true)

	// string
	h().Field(&m.Name).EQ("alice")

	// time.Time
	h().Field(&m.CreatedAt).LT(time.Now())

	// Pointer field: T, *T, typed-nil, untyped-nil — все допустимы.
	s := "x"
	h().Field(&m.Email).EQ("a")
	h().Field(&m.Email).EQ(&s)
	h().Field(&m.Email).EQ((*string)(nil))
	h().Field(&m.Email).EQ(nil)
}

type Age int
type UserID int

type ModelNamed struct {
	Age    Age
	UserID UserID
}

func namedPositives() {
	m := &ModelNamed{}
	var a Age = 20
	h().Field(&m.Age).EQ(18) // untyped int → Age
	h().Field(&m.Age).EQ(a)  // named var of same type
}

func negatives() {
	m := &Model{}

	h().Field(&m.Age).EQ("18")   // want `GPL001: EQ: argument type string is not compatible with field type int`
	h().Field(&m.Age).EQ(3.14)   // want `GPL001: EQ: argument type float64 is not compatible with field type int`
	h().Field(&m.Age).EQ(nil)    // want `GPL001: EQ: argument type untyped nil is not compatible with field type int`
	h().Field(&m.Age).GTE(true)  // want `GPL001: GTE: argument type bool is not compatible with field type int`
	h().Field(&m.Name).EQ(18)    // want `GPL001: EQ: argument type int is not compatible with field type string`
	h().Field(&m.Email).EQ(42)   // want `GPL001: EQ: argument type int is not compatible with field type \*string`
	h().Field(&m.CreatedAt).EQ("2026-01-01") // want `GPL001: EQ: argument type string is not compatible with field type time\.Time`
}

func namedNegatives() {
	m := &ModelNamed{}
	// Age and UserID have the same underlying int, but are distinct named types.
	h().Field(&m.Age).EQ(UserID(1)) // want `GPL001: EQ: argument type compat\.UserID is not compatible with field type compat\.Age`
}
