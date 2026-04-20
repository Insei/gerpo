package stringonly

import (
	"time"

	"github.com/insei/gerpo/types"
)

type Alias string

type Model struct {
	Name      string
	Email     *string
	Slug      Alias
	Age       int
	Active    bool
	CreatedAt time.Time
}

type whereTarget struct{}

func (whereTarget) Field(_ any) types.WhereOperation            { return nil }
func (whereTarget) Group(_ func(types.WhereTarget)) types.ANDOR { return nil }

func h() types.WhereTarget { return whereTarget{} }

func positives() {
	m := &Model{}

	h().Field(&m.Name).Contains("a")
	h().Field(&m.Name).NotContains("a")
	h().Field(&m.Name).StartsWith("a")
	h().Field(&m.Name).NotStartsWith("a")
	h().Field(&m.Name).EndsWith("a")
	h().Field(&m.Name).NotEndsWith("a")

	h().Field(&m.Email).StartsWith("a@")
	h().Field(&m.Email).Contains("@")

	h().Field(&m.Slug).Contains("x")

	h().Field(&m.Name).EQFold("ALICE")
	h().Field(&m.Name).NotEQFold("ALICE")
	h().Field(&m.Name).ContainsFold("AL")
	h().Field(&m.Name).NotContainsFold("AL")
	h().Field(&m.Name).StartsWithFold("AL")
	h().Field(&m.Name).NotStartsWithFold("AL")
	h().Field(&m.Name).EndsWithFold("CE")
	h().Field(&m.Name).NotEndsWithFold("CE")
	h().Field(&m.Email).EQFold("ALICE")
}

func likeNegatives() {
	m := &Model{}

	h().Field(&m.Age).Contains("x")          // want `GPL003: Contains is only applicable to string/\*string fields, got int`
	h().Field(&m.Active).NotContains("x")    // want `GPL003: NotContains is only applicable to string/\*string fields, got bool`
	h().Field(&m.CreatedAt).StartsWith("x")  // want `GPL003: StartsWith is only applicable to string/\*string fields, got time\.Time`
	h().Field(&m.Age).NotStartsWith("x")     // want `GPL003: NotStartsWith is only applicable to string/\*string fields, got int`
	h().Field(&m.Age).EndsWith("x")          // want `GPL003: EndsWith is only applicable to string/\*string fields, got int`
	h().Field(&m.Age).NotEndsWith("x")       // want `GPL003: NotEndsWith is only applicable to string/\*string fields, got int`
}

func foldNegatives() {
	m := &Model{}

	h().Field(&m.Age).EQFold("x")            // want `GPL003: EQFold is only applicable to string/\*string fields, got int`
	h().Field(&m.Age).NotEQFold("x")         // want `GPL003: NotEQFold is only applicable to string/\*string fields, got int`
	h().Field(&m.Age).ContainsFold("x")      // want `GPL003: ContainsFold is only applicable to string/\*string fields, got int`
	h().Field(&m.Age).NotContainsFold("x")   // want `GPL003: NotContainsFold is only applicable to string/\*string fields, got int`
	h().Field(&m.Age).StartsWithFold("x")    // want `GPL003: StartsWithFold is only applicable to string/\*string fields, got int`
	h().Field(&m.Age).NotStartsWithFold("x") // want `GPL003: NotStartsWithFold is only applicable to string/\*string fields, got int`
	h().Field(&m.Age).EndsWithFold("x")      // want `GPL003: EndsWithFold is only applicable to string/\*string fields, got int`
	h().Field(&m.Age).NotEndsWithFold("x")   // want `GPL003: NotEndsWithFold is only applicable to string/\*string fields, got int`
}
