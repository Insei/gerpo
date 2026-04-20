package gerpolint

type operatorKind int

const (
	opScalar operatorKind = iota
	opVariadic
)

type operatorCategory int

const (
	catAny operatorCategory = iota
	catStringOnly
)

type operatorSpec struct {
	name     string
	kind     operatorKind
	category operatorCategory
}

// operators mirrors the methods defined on types.WhereOperation in
// github.com/insei/gerpo/types. The analyzer looks up called methods in this
// table; unknown names are ignored.
var operators = map[string]operatorSpec{
	"EQ":    {name: "EQ", kind: opScalar, category: catAny},
	"NotEQ": {name: "NotEQ", kind: opScalar, category: catAny},
	"LT":    {name: "LT", kind: opScalar, category: catAny},
	"LTE":   {name: "LTE", kind: opScalar, category: catAny},
	"GT":    {name: "GT", kind: opScalar, category: catAny},
	"GTE":   {name: "GTE", kind: opScalar, category: catAny},

	"In":    {name: "In", kind: opVariadic, category: catAny},
	"NotIn": {name: "NotIn", kind: opVariadic, category: catAny},

	"Contains":      {name: "Contains", kind: opScalar, category: catStringOnly},
	"NotContains":   {name: "NotContains", kind: opScalar, category: catStringOnly},
	"StartsWith":    {name: "StartsWith", kind: opScalar, category: catStringOnly},
	"NotStartsWith": {name: "NotStartsWith", kind: opScalar, category: catStringOnly},
	"EndsWith":      {name: "EndsWith", kind: opScalar, category: catStringOnly},
	"NotEndsWith":   {name: "NotEndsWith", kind: opScalar, category: catStringOnly},

	"EQFold":            {name: "EQFold", kind: opScalar, category: catStringOnly},
	"NotEQFold":         {name: "NotEQFold", kind: opScalar, category: catStringOnly},
	"ContainsFold":      {name: "ContainsFold", kind: opScalar, category: catStringOnly},
	"NotContainsFold":   {name: "NotContainsFold", kind: opScalar, category: catStringOnly},
	"StartsWithFold":    {name: "StartsWithFold", kind: opScalar, category: catStringOnly},
	"NotStartsWithFold": {name: "NotStartsWithFold", kind: opScalar, category: catStringOnly},
	"EndsWithFold":      {name: "EndsWithFold", kind: opScalar, category: catStringOnly},
	"NotEndsWithFold":   {name: "NotEndsWithFold", kind: opScalar, category: catStringOnly},
}
