package filters

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insei/fmap/v3"

	"github.com/insei/gerpo/types"
)

// fixture exercises the dispatch table: every primitive kind we care about plus
// time.Time / uuid.UUID, plus pointer wrappers that should resolve through the
// dereference branch.
type fixture struct {
	Bool      bool
	BoolPtr   *bool
	Str       string
	StrPtr    *string
	Int       int
	Int8      int8
	Int16     int16
	Int32     int32
	Int64     int64
	Uint      uint
	Uint8     uint8
	Uint16    uint16
	Uint32    uint32
	Uint64    uint64
	Float32   float32
	Float64   float64
	IntPtr    *int
	Time      time.Time
	TimePtr   *time.Time
	UUID      uuid.UUID
	UUIDPtr   *uuid.UUID
	Status    fixtureStatus
	StatusPtr *fixtureStatus
}

type fixtureStatus string

func fieldOf(t *testing.T, name string) fmap.Field {
	t.Helper()
	store, err := fmap.GetFrom(&fixture{})
	if err != nil {
		t.Fatalf("fmap: %v", err)
	}
	f, ok := store.Find(name)
	if !ok {
		t.Fatalf("fmap: field %s not found", name)
	}
	return f
}

func opSet(filters map[types.Operation]Filter) []string {
	out := make([]string, 0, len(filters))
	for op := range filters {
		out = append(out, string(op))
	}
	sort.Strings(out)
	return out
}

func TestDefaults_PrimitivesMatchLegacyMapping(t *testing.T) {
	r := newRegistry()

	cases := []struct {
		field    string
		expected []types.Operation
	}{
		{"Bool", []types.Operation{types.OperationEQ, types.OperationNotEQ}},
		{"Str", []types.Operation{
			types.OperationEQ, types.OperationNotEQ,
			types.OperationIn, types.OperationNotIn,
			types.OperationContains, types.OperationNotContains,
			types.OperationStartsWith, types.OperationNotStartsWith,
			types.OperationEndsWith, types.OperationNotEndsWith,
			types.OperationEQFold, types.OperationNotEQFold,
			types.OperationContainsFold, types.OperationNotContainsFold,
			types.OperationStartsWithFold, types.OperationNotStartsWithFold,
			types.OperationEndsWithFold, types.OperationNotEndsWithFold,
		}},
		{"Int", numericOps()},
		{"Int8", numericOps()},
		{"Int16", numericOps()},
		{"Int32", numericOps()},
		{"Int64", numericOps()},
		{"Uint", numericOps()},
		{"Uint8", numericOps()},
		{"Uint16", numericOps()},
		{"Uint32", numericOps()},
		{"Uint64", numericOps()},
		{"Float32", numericOps()},
		{"Float64", numericOps()},
		{"Time", []types.Operation{types.OperationLT, types.OperationLTE, types.OperationGT, types.OperationGTE}},
		{"UUID", []types.Operation{types.OperationEQ, types.OperationNotEQ, types.OperationIn, types.OperationNotIn}},
	}
	for _, c := range cases {
		t.Run(c.field, func(t *testing.T) {
			got := r.Apply(fieldOf(t, c.field), c.field)
			want := opsAsStrings(c.expected)
			if !equalStringSets(opSet(got), want) {
				t.Fatalf("ops mismatch: got %v want %v", opSet(got), want)
			}
		})
	}
}

// PtrFieldsAddNullableEquality verifies that pointer wrappers around supported
// types pick up EQ/NotEQ on top of the dereferenced kind's filters — needed
// for IS NULL / IS NOT NULL semantics.
func TestDefaults_PtrFieldsAddNullableEquality(t *testing.T) {
	r := newRegistry()

	intPtrOps := opSet(r.Apply(fieldOf(t, "IntPtr"), "users.int_ptr"))
	wantNumeric := opsAsStrings(numericOps())
	if !equalStringSets(intPtrOps, wantNumeric) {
		t.Fatalf("IntPtr ops mismatch: got %v want superset of %v", intPtrOps, wantNumeric)
	}

	timePtrOps := opSet(r.Apply(fieldOf(t, "TimePtr"), "users.time_ptr"))
	wantTime := opsAsStrings([]types.Operation{
		types.OperationEQ, types.OperationNotEQ, // added by the ptr branch
		types.OperationLT, types.OperationLTE, types.OperationGT, types.OperationGTE,
	})
	if !equalStringSets(timePtrOps, wantTime) {
		t.Fatalf("TimePtr ops mismatch: got %v want %v", timePtrOps, wantTime)
	}

	boolPtrOps := opSet(r.Apply(fieldOf(t, "BoolPtr"), "users.bool_ptr"))
	wantBool := opsAsStrings([]types.Operation{types.OperationEQ, types.OperationNotEQ})
	if !equalStringSets(boolPtrOps, wantBool) {
		t.Fatalf("BoolPtr ops mismatch: got %v want %v", boolPtrOps, wantBool)
	}
}

func TestRegister_CustomTypeBeatsKindBucket(t *testing.T) {
	r := newRegistry()

	// Without registration: Status falls through to the String bucket and gets
	// the full string operator list.
	got := opSet(r.Apply(fieldOf(t, "Status"), "users.status"))
	if !contains(got, string(types.OperationContains)) {
		t.Fatalf("expected Status to inherit String ops by default, got %v", got)
	}

	// After registering Status with a narrow set, custom registration wins.
	r.Register(fixtureStatus("")).
		Allow(types.OperationEQ, types.OperationIn)

	got = opSet(r.Apply(fieldOf(t, "Status"), "users.status"))
	want := opsAsStrings([]types.Operation{types.OperationEQ, types.OperationIn})
	if !equalStringSets(got, want) {
		t.Fatalf("after Register: got %v want %v", got, want)
	}

	// Lookup returns the bucket; types not registered return nil.
	if r.Lookup(reflect.TypeOf(fixtureStatus(""))) == nil {
		t.Fatal("Lookup returned nil for a registered type")
	}
	if r.Lookup(reflect.TypeOf("plain string")) != nil {
		t.Fatal("Lookup returned non-nil for an unregistered type")
	}
}

func TestOverride_ReplacesStockSQL(t *testing.T) {
	r := newRegistry()

	r.Time.Override(types.OperationEQ, Bound{SQL: "TRUNC(? AS date) = TRUNC(? AS date)"})

	got := r.Apply(fieldOf(t, "Time"), "users.created_at")
	fn, ok := got[types.OperationEQ]
	if !ok {
		t.Fatal("Override(EQ) should add EQ to allowed set")
	}
	sql, args, err := fn(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	if sql != "TRUNC(? AS date) = TRUNC(? AS date)" {
		t.Fatalf("override SQL not applied: %q", sql)
	}
	// Bound{} compiles to one bound arg per call (the user value).
	if len(args) != 1 {
		t.Fatalf("expected 1 bound arg, got %d", len(args))
	}

	// LT/GT/LTE/GTE remain stock.
	if _, ok := got[types.OperationLT]; !ok {
		t.Fatal("Override(EQ) must not strip the rest of Time defaults")
	}
}

func TestRemove_DropsOpFromBucket(t *testing.T) {
	r := newRegistry()
	r.Numeric.Remove(types.OperationLT)

	got := opSet(r.Apply(fieldOf(t, "Int"), "users.int"))
	if contains(got, string(types.OperationLT)) {
		t.Fatalf("Remove(LT) should drop LT from Int filters: %v", got)
	}
}

func TestApply_UnknownTypeReturnsEmpty(t *testing.T) {
	type unknown struct{ X int }
	type holder struct{ U unknown }
	store, err := fmap.GetFrom(&holder{})
	if err != nil {
		t.Fatalf("fmap: %v", err)
	}
	f, ok := store.Find("U")
	if !ok {
		t.Fatal("U not found")
	}
	r := newRegistry()
	got := r.Apply(f, "x.u")
	if len(got) != 0 {
		t.Fatalf("expected empty filter map for unknown struct type, got %v", opSet(got))
	}
}

func numericOps() []types.Operation {
	return []types.Operation{
		types.OperationEQ, types.OperationNotEQ,
		types.OperationLT, types.OperationLTE,
		types.OperationGT, types.OperationGTE,
		types.OperationIn, types.OperationNotIn,
	}
}

func opsAsStrings(ops []types.Operation) []string {
	out := make([]string, len(ops))
	for i, op := range ops {
		out[i] = string(op)
	}
	sort.Strings(out)
	return out
}

func equalStringSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
