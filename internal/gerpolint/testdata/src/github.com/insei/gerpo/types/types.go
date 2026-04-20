// Package types is a stub mirror of github.com/insei/gerpo/types, exposing
// only the interfaces the analyzer inspects: WhereTarget, WhereOperation,
// and ANDOR. The analyzer identifies gerpo calls by pkg path, so the
// directory structure under testdata/src is significant.
package types

type WhereOperation interface {
	EQ(val any) ANDOR
	NotEQ(val any) ANDOR
	LT(val any) ANDOR
	LTE(val any) ANDOR
	GT(val any) ANDOR
	GTE(val any) ANDOR

	In(vals ...any) ANDOR
	NotIn(vals ...any) ANDOR

	Contains(val any) ANDOR
	NotContains(val any) ANDOR
	StartsWith(val any) ANDOR
	NotStartsWith(val any) ANDOR
	EndsWith(val any) ANDOR
	NotEndsWith(val any) ANDOR

	EQFold(val any) ANDOR
	NotEQFold(val any) ANDOR
	ContainsFold(val any) ANDOR
	NotContainsFold(val any) ANDOR
	StartsWithFold(val any) ANDOR
	NotStartsWithFold(val any) ANDOR
	EndsWithFold(val any) ANDOR
	NotEndsWithFold(val any) ANDOR
}

type WhereTarget interface {
	Field(fieldPtr any) WhereOperation
	Group(fn func(t WhereTarget)) ANDOR
}

type ANDOR interface {
	AND() WhereTarget
	OR() WhereTarget
}
