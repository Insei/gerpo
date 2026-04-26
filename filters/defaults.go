package filters

import (
	"reflect"
	"time"

	"github.com/google/uuid"

	"github.com/insei/gerpo/types"
)

// newRegistry returns a registry pre-populated with the operator sets that
// gerpo ships out of the box for primitive Go types, time.Time and uuid.UUID.
// Built-in buckets are exported on the registry so users can inspect or
// extend them via the Registry singleton.
func newRegistry() *registry {
	r := &registry{
		custom: map[reflect.Type]*TypeBucket{},
	}

	// Bool: equality only.
	r.Bool = &KindBucket{kinds: []reflect.Kind{reflect.Bool}}
	r.Bool.Allow(types.OperationEQ, types.OperationNotEQ)

	// String: equality, set membership, full LIKE / fold suite.
	r.String = &KindBucket{kinds: []reflect.Kind{reflect.String}}
	r.String.Allow(
		types.OperationEQ, types.OperationNotEQ,
		types.OperationIn, types.OperationNotIn,
		types.OperationContains, types.OperationNotContains,
		types.OperationStartsWith, types.OperationNotStartsWith,
		types.OperationEndsWith, types.OperationNotEndsWith,
		types.OperationEQFold, types.OperationNotEQFold,
		types.OperationContainsFold, types.OperationNotContainsFold,
		types.OperationStartsWithFold, types.OperationNotStartsWithFold,
		types.OperationEndsWithFold, types.OperationNotEndsWithFold,
	)

	// Numeric: equality, ordering, set membership.
	r.Numeric = &KindBucket{kinds: []reflect.Kind{
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
	}}
	r.Numeric.Allow(
		types.OperationEQ, types.OperationNotEQ,
		types.OperationLT, types.OperationLTE,
		types.OperationGT, types.OperationGTE,
		types.OperationIn, types.OperationNotIn,
	)

	// time.Time: ordering only — equality on timestamps almost always wants
	// range/window, so the historical default left it out.
	r.Time = &TypeBucket{rt: reflect.TypeOf(time.Time{})}
	r.Time.Allow(
		types.OperationLT, types.OperationLTE,
		types.OperationGT, types.OperationGTE,
	)

	// uuid.UUID: equality and set membership; ordering is meaningless.
	r.UUID = &TypeBucket{rt: reflect.TypeOf(uuid.UUID{})}
	r.UUID.Allow(
		types.OperationEQ, types.OperationNotEQ,
		types.OperationIn, types.OperationNotIn,
	)

	return r
}
