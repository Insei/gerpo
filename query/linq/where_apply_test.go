package linq

import (
	"context"
	"fmt"
	"testing"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeWhere is a sqlpart.Where double that records every structural call.
type fakeWhere struct {
	sqlpart.Where
	calls []string
}

func (f *fakeWhere) StartGroup() { f.calls = append(f.calls, "StartGroup") }
func (f *fakeWhere) EndGroup()   { f.calls = append(f.calls, "EndGroup") }
func (f *fakeWhere) AND()        { f.calls = append(f.calls, "AND") }
func (f *fakeWhere) OR()         { f.calls = append(f.calls, "OR") }
func (f *fakeWhere) AppendCondition(col types.Column, op types.Operation, _ any) error {
	sql, _ := col.Name()
	f.calls = append(f.calls, "cond:"+sql+"/"+string(op))
	return nil
}

type whereApplier struct {
	storage types.ColumnsStorage
	where   sqlpart.Where
}

func (a *whereApplier) ColumnsStorage() types.ColumnsStorage { return a.storage }
func (a *whereApplier) Where() sqlpart.Where                 { return a.where }

type failWhere struct{ sqlpart.Where }

func (failWhere) StartGroup() {}
func (failWhere) EndGroup()   {}
func (failWhere) AppendCondition(types.Column, types.Operation, any) error {
	return fmt.Errorf("cond fail")
}

func TestWhereBuilder_IsEmpty(t *testing.T) {
	b := NewWhereBuilder(nil)
	assert.True(t, b.IsEmpty())
	b.AND()
	assert.False(t, b.IsEmpty())
}

func TestWhereBuilder_Apply_Empty_Noop(t *testing.T) {
	b := NewWhereBuilder(nil)
	require.NoError(t, b.Apply(&whereApplier{where: &fakeWhere{}}))
}

func TestWhereBuilder_Column_AllOperators(t *testing.T) {
	col := &mockColumn{name: "x", hasName: true}
	b := NewWhereBuilder(nil)

	// Touch every chainable operator — just verifies they build a non-empty plan.
	b.Column(col).EQ(1)
	b.Column(col).NotEQ(1)
	b.Column(col).GT(1)
	b.Column(col).GTE(1)
	b.Column(col).LT(1)
	b.Column(col).LTE(1)
	b.Column(col).In(1, 2)
	b.Column(col).NotIn(1, 2)
	b.Column(col).Contains("a")
	b.Column(col).NotContains("a")
	b.Column(col).StartsWith("a")
	b.Column(col).NotStartsWith("a")
	b.Column(col).EndsWith("a")
	b.Column(col).NotEndsWith("a")
	b.Column(col).EQFold("a")
	b.Column(col).NotEQFold("a")
	b.Column(col).ContainsFold("a")
	b.Column(col).NotContainsFold("a")
	b.Column(col).StartsWithFold("a")
	b.Column(col).NotStartsWithFold("a")
	b.Column(col).EndsWithFold("a")
	b.Column(col).NotEndsWithFold("a")

	w := &fakeWhere{}
	require.NoError(t, b.Apply(&whereApplier{where: w}))
	// Should have a StartGroup + 15 conditions (with implicit ANDs) + EndGroup.
	assert.Contains(t, w.calls, "StartGroup")
	assert.Contains(t, w.calls, "EndGroup")
}

func TestWhereBuilder_Field_ResolvesColumn(t *testing.T) {
	type m struct{ Name string }
	model := &m{}
	col := &mockColumn{name: "name", hasName: true}
	storage := &mockColumnsStorage{columns: map[any]types.Column{&model.Name: col}}

	b := NewWhereBuilder(model)
	b.Field(&model.Name).EQ("alice")

	w := &fakeWhere{}
	require.NoError(t, b.Apply(&whereApplier{storage: storage, where: w}))
	assert.Contains(t, w.calls, "cond:name/eq")
}

func TestWhereBuilder_Field_ResolveError(t *testing.T) {
	type m struct{ Name string }
	model := &m{}
	storage := &mockColumnsStorage{columns: map[any]types.Column{}}

	b := NewWhereBuilder(model)
	b.Field(&model.Name).EQ("x")

	err := b.Apply(&whereApplier{storage: storage, where: &fakeWhere{}})
	require.Error(t, err)
}

func TestWhereBuilder_Column_NilRejected(t *testing.T) {
	b := NewWhereBuilder(nil)
	b.Column(nil).EQ("x")
	err := b.Apply(&whereApplier{where: &fakeWhere{}})
	require.Error(t, err)
}

func TestWhereBuilder_Group_WrapsWithStartEnd(t *testing.T) {
	col := &mockColumn{name: "name", hasName: true}
	b := NewWhereBuilder(nil)
	b.Group(func(t types.WhereTarget) {
		t.Column(col).EQ("a").OR().Column(col).EQ("b")
	}).AND().Column(col).EQ("c")

	w := &fakeWhere{}
	require.NoError(t, b.Apply(&whereApplier{where: w}))
	// Outer StartGroup + inner StartGroup + EndGroup + trailing ops + EndGroup.
	startCount := 0
	endCount := 0
	for _, c := range w.calls {
		switch c {
		case "StartGroup":
			startCount++
		case "EndGroup":
			endCount++
		}
	}
	assert.Equal(t, 2, startCount)
	assert.Equal(t, 2, endCount)
}

func TestWhereBuilder_Apply_ConditionErrorPropagates(t *testing.T) {
	col := &mockColumn{name: "x", hasName: true}
	b := NewWhereBuilder(nil)
	b.Column(col).EQ(1)

	err := b.Apply(&whereApplier{where: failWhere{}})
	require.Error(t, err)
}

var _ = context.Background // keep context import — fakeWhere may grow
