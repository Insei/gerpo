package filters

import (
	"context"
	"reflect"
	"sync"

	"github.com/insei/fmap/v3"

	"github.com/insei/gerpo/internal/sqltpl"
	"github.com/insei/gerpo/types"
)

// Filter is the args-based callback shape stored on a column's
// SQLFilterManager. It mirrors types.SQLFilterManager.AddFilterFnArgs.
type Filter func(ctx context.Context, value any) (sql string, args []any, err error)

// stockFilter returns the standard sqltpl-based filter for an operation. It is
// used by buckets to fulfill Allow(); Override() injects user-supplied SQL via
// CompileFilter instead.
func stockFilter(op types.Operation, columnSQL string) Filter {
	if gen := stockGenerator(op); gen != nil {
		legacy := gen(columnSQL)
		return func(_ context.Context, value any) (string, []any, error) {
			sql, appendValue := legacy(nil, value)
			if !appendValue {
				return sql, nil, nil
			}
			return sql, []any{value}, nil
		}
	}
	return nil
}

// stockGenerator maps every supported operator to the corresponding sqltpl
// helper. Operators without a stock template (because the user is expected to
// supply one via Override) yield nil.
func stockGenerator(op types.Operation) sqltpl.Generator {
	switch op {
	case types.OperationEQ:
		return sqltpl.EQ
	case types.OperationNotEQ:
		return sqltpl.NotEQ
	case types.OperationLT:
		return sqltpl.LT
	case types.OperationLTE:
		return sqltpl.LTE
	case types.OperationGT:
		return sqltpl.GT
	case types.OperationGTE:
		return sqltpl.GTE
	case types.OperationIn:
		return sqltpl.In
	case types.OperationNotIn:
		return sqltpl.NotIn
	case types.OperationContains:
		return sqltpl.Contains
	case types.OperationNotContains:
		return sqltpl.NotContains
	case types.OperationStartsWith:
		return sqltpl.StartsWith
	case types.OperationNotStartsWith:
		return sqltpl.NotStartsWith
	case types.OperationEndsWith:
		return sqltpl.EndsWith
	case types.OperationNotEndsWith:
		return sqltpl.NotEndsWith
	case types.OperationEQFold:
		return sqltpl.EQFold
	case types.OperationNotEQFold:
		return sqltpl.NotEQFold
	case types.OperationContainsFold:
		return sqltpl.ContainsFold
	case types.OperationNotContainsFold:
		return sqltpl.NotContainsFold
	case types.OperationStartsWithFold:
		return sqltpl.StartsWithFold
	case types.OperationNotStartsWithFold:
		return sqltpl.NotStartsWithFold
	case types.OperationEndsWithFold:
		return sqltpl.EndsWithFold
	case types.OperationNotEndsWithFold:
		return sqltpl.NotEndsWithFold
	}
	return nil
}

// TypeBucket holds the set of operators allowed for one concrete reflect.Type
// plus any user overrides.
type TypeBucket struct {
	mu        sync.RWMutex
	rt        reflect.Type
	allowed   []types.Operation
	overrides map[types.Operation]FilterSpec
}

// Allow registers ops with the stock SQL templates. Repeated ops are
// deduplicated. An Allow on an op that has an Override removes the override —
// switch back to the default template.
func (b *TypeBucket) Allow(ops ...types.Operation) *TypeBucket {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, op := range ops {
		if !containsOp(b.allowed, op) {
			b.allowed = append(b.allowed, op)
		}
		delete(b.overrides, op)
	}
	return b
}

// Override registers a user-supplied SQL fragment for op. The op is implicitly
// allowed.
func (b *TypeBucket) Override(op types.Operation, spec FilterSpec) *TypeBucket {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !containsOp(b.allowed, op) {
		b.allowed = append(b.allowed, op)
	}
	if b.overrides == nil {
		b.overrides = map[types.Operation]FilterSpec{}
	}
	b.overrides[op] = spec
	return b
}

// Remove withdraws ops from the bucket — both the Allow entry and any Override.
func (b *TypeBucket) Remove(ops ...types.Operation) *TypeBucket {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, op := range ops {
		b.allowed = removeOp(b.allowed, op)
		delete(b.overrides, op)
	}
	return b
}

// Operations returns a copy of the currently allowed operators — useful for
// inspection in tests and self-checks.
func (b *TypeBucket) Operations() []types.Operation {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]types.Operation, len(b.allowed))
	copy(out, b.allowed)
	return out
}

func (b *TypeBucket) fillInto(out map[types.Operation]Filter, columnSQL string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, op := range b.allowed {
		if spec, ok := b.overrides[op]; ok {
			out[op] = CompileFilter(spec)
			continue
		}
		if f := stockFilter(op, columnSQL); f != nil {
			out[op] = f
		}
	}
}

// KindBucket holds the operator set shared by every reflect.Kind in kinds.
// Used for primitive families (Numeric covers all int/uint/float kinds).
type KindBucket struct {
	mu        sync.RWMutex
	kinds     []reflect.Kind
	allowed   []types.Operation
	overrides map[types.Operation]FilterSpec
}

func (b *KindBucket) Allow(ops ...types.Operation) *KindBucket {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, op := range ops {
		if !containsOp(b.allowed, op) {
			b.allowed = append(b.allowed, op)
		}
		delete(b.overrides, op)
	}
	return b
}

func (b *KindBucket) Override(op types.Operation, spec FilterSpec) *KindBucket {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !containsOp(b.allowed, op) {
		b.allowed = append(b.allowed, op)
	}
	if b.overrides == nil {
		b.overrides = map[types.Operation]FilterSpec{}
	}
	b.overrides[op] = spec
	return b
}

func (b *KindBucket) Remove(ops ...types.Operation) *KindBucket {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, op := range ops {
		b.allowed = removeOp(b.allowed, op)
		delete(b.overrides, op)
	}
	return b
}

func (b *KindBucket) Operations() []types.Operation {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]types.Operation, len(b.allowed))
	copy(out, b.allowed)
	return out
}

func (b *KindBucket) matches(k reflect.Kind) bool {
	for _, kk := range b.kinds {
		if kk == k {
			return true
		}
	}
	return false
}

func (b *KindBucket) fillInto(out map[types.Operation]Filter, columnSQL string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, op := range b.allowed {
		if spec, ok := b.overrides[op]; ok {
			out[op] = CompileFilter(spec)
			continue
		}
		if f := stockFilter(op, columnSQL); f != nil {
			out[op] = f
		}
	}
}

// Registry is the project-wide filter registry. The single value lives in the
// exported var Registry; users mutate it once at process start (typically in
// init()) and gerpo consults it whenever a column is built.
type registry struct {
	mu sync.RWMutex

	Bool    *KindBucket
	String  *KindBucket
	Numeric *KindBucket
	Time    *TypeBucket
	UUID    *TypeBucket

	custom map[reflect.Type]*TypeBucket
}

// Register binds a TypeBucket to the type of example. Subsequent calls with
// the same type return the existing bucket — Allow/Override are additive.
//
// example is used purely for reflect.TypeOf(example); pass a zero value
// (Money{}, Status("")). The stored type is the dereferenced one — pointer
// fields find their custom registration through the deref step in Apply.
func (r *registry) Register(example any) *TypeBucket {
	rt := reflect.TypeOf(example)
	for rt != nil && rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.custom[rt]; ok {
		return b
	}
	b := &TypeBucket{rt: rt}
	r.custom[rt] = b
	return b
}

// Lookup returns the bucket registered for t, or nil if nothing is registered.
// Inspection helper for tests and self-diagnostic code.
func (r *registry) Lookup(t reflect.Type) *TypeBucket {
	if t == nil {
		return nil
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.custom[t]
}

// Apply resolves the filter set for (field, columnSQL). Resolution order:
//  1. pointer fields receive stock EQ/NotEQ for NULL handling, then deref;
//  2. custom-registered reflect.Type wins over kind buckets;
//  3. time.Time and uuid.UUID use the named TypeBuckets;
//  4. primitive kinds fall through to Bool/String/Numeric KindBuckets;
//  5. unknown types return an empty map — callers decide how to handle that.
func (r *registry) Apply(field fmap.Field, columnSQL string) map[types.Operation]Filter {
	out := map[types.Operation]Filter{}
	if field == nil {
		return out
	}

	// 1. ptr → stock EQ/NotEQ for IS NULL / IS NOT NULL semantics.
	if field.GetType().Kind() == reflect.Ptr {
		if f := stockFilter(types.OperationEQ, columnSQL); f != nil {
			out[types.OperationEQ] = f
		}
		if f := stockFilter(types.OperationNotEQ, columnSQL); f != nil {
			out[types.OperationNotEQ] = f
		}
	}

	deref := field.GetDereferencedType()
	if deref == nil {
		return out
	}

	// 2. custom type registration wins.
	r.mu.RLock()
	cb, hasCustom := r.custom[deref]
	r.mu.RUnlock()
	if hasCustom {
		cb.fillInto(out, columnSQL)
		return out
	}

	// 3. named built-in buckets — match by exact reflect.Type.
	switch {
	case r.Time != nil && deref == r.Time.rt:
		r.Time.fillInto(out, columnSQL)
		return out
	case r.UUID != nil && deref == r.UUID.rt:
		r.UUID.fillInto(out, columnSQL)
		return out
	}

	// 4. primitive Kind buckets.
	kind := deref.Kind()
	switch {
	case r.Bool != nil && r.Bool.matches(kind):
		r.Bool.fillInto(out, columnSQL)
	case r.String != nil && r.String.matches(kind):
		r.String.fillInto(out, columnSQL)
	case r.Numeric != nil && r.Numeric.matches(kind):
		r.Numeric.fillInto(out, columnSQL)
	}
	return out
}

// Registry is the global, mutable instance. Mutate during init(); reads happen
// at column construction time inside gerpo.
var Registry = newRegistry()

// Snapshot returns a function that restores the global Registry to the state
// it had at the moment Snapshot was called. Intended for tests that mutate
// the registry — pair with t.Cleanup so the next test starts from a clean
// slate. Mutations include Register/Unregister, bucket Allow/Override/Remove.
func Snapshot() (restore func()) {
	r := Registry
	r.mu.Lock()
	bool_ := cloneKindBucket(r.Bool)
	str_ := cloneKindBucket(r.String)
	num_ := cloneKindBucket(r.Numeric)
	time_ := cloneTypeBucket(r.Time)
	uuid_ := cloneTypeBucket(r.UUID)
	custom_ := make(map[reflect.Type]*TypeBucket, len(r.custom))
	for k, v := range r.custom {
		custom_[k] = cloneTypeBucket(v)
	}
	r.mu.Unlock()
	return func() {
		r.mu.Lock()
		r.Bool = bool_
		r.String = str_
		r.Numeric = num_
		r.Time = time_
		r.UUID = uuid_
		r.custom = custom_
		r.mu.Unlock()
	}
}

func cloneKindBucket(b *KindBucket) *KindBucket {
	if b == nil {
		return nil
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := &KindBucket{
		kinds:   append([]reflect.Kind(nil), b.kinds...),
		allowed: append([]types.Operation(nil), b.allowed...),
	}
	if b.overrides != nil {
		out.overrides = make(map[types.Operation]FilterSpec, len(b.overrides))
		for op, spec := range b.overrides {
			out.overrides[op] = spec
		}
	}
	return out
}

func cloneTypeBucket(b *TypeBucket) *TypeBucket {
	if b == nil {
		return nil
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := &TypeBucket{
		rt:      b.rt,
		allowed: append([]types.Operation(nil), b.allowed...),
	}
	if b.overrides != nil {
		out.overrides = make(map[types.Operation]FilterSpec, len(b.overrides))
		for op, spec := range b.overrides {
			out.overrides[op] = spec
		}
	}
	return out
}

func containsOp(s []types.Operation, op types.Operation) bool {
	for _, o := range s {
		if o == op {
			return true
		}
	}
	return false
}

func removeOp(s []types.Operation, op types.Operation) []types.Operation {
	for i, o := range s {
		if o == op {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
