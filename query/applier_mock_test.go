package query

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

// --- Column / storage / exec-columns fakes ------------------------------------

type mockColumn struct {
	types.Column
	name    string
	hasName bool
	sql     string
	allowed bool
	filters map[types.Operation]func(ctx context.Context, value any) (string, []any, error)
}

func (m *mockColumn) IsAllowedAction(types.SQLAction) bool { return m.allowed }
func (m *mockColumn) ToSQL(context.Context) string         { return m.sql }
func (m *mockColumn) Name() (string, bool)                 { return m.name, m.hasName }

func (m *mockColumn) GetFilterFn(op types.Operation) (func(ctx context.Context, value any) (string, []any, error), bool) {
	fn, ok := m.filters[op]
	return fn, ok
}

type mockColumnsStorage struct {
	types.ColumnsStorage
	byPtr   map[any]types.Column
	execCol types.ExecutionColumns
	newErr  error
}

func (m *mockColumnsStorage) GetByFieldPtr(_ any, fieldPtr any) (types.Column, error) {
	if col, ok := m.byPtr[fieldPtr]; ok {
		return col, nil
	}
	return nil, fmt.Errorf("column not found for %v", fieldPtr)
}

func (m *mockColumnsStorage) NewExecutionColumns(context.Context, types.SQLAction) types.ExecutionColumns {
	return m.execCol
}

type mockExecCols struct {
	types.ExecutionColumns
	all      []types.Column
	excluded []types.Column
	only     []types.Column
}

func (m *mockExecCols) GetAll() []types.Column { return m.all }
func (m *mockExecCols) Exclude(cols ...types.Column) {
	m.excluded = append(m.excluded, cols...)
}
func (m *mockExecCols) Only(cols ...types.Column) { m.only = cols }
func (m *mockExecCols) GetByFieldPtr(model, fieldPtr any) (types.Column, error) {
	return nil, fmt.Errorf("not implemented")
}

// --- sqlpart fakes -----------------------------------------------------------

type mockWhere struct {
	sqlpart.Where
	fragments []string
	values    []any
}

func (m *mockWhere) StartGroup() { m.fragments = append(m.fragments, "(") }
func (m *mockWhere) EndGroup()   { m.fragments = append(m.fragments, ")") }
func (m *mockWhere) AND()        { m.fragments = append(m.fragments, "AND") }
func (m *mockWhere) OR()         { m.fragments = append(m.fragments, "OR") }
func (m *mockWhere) AppendSQLWithValues(sql string, _ bool, _ any) {
	m.fragments = append(m.fragments, sql)
}
func (m *mockWhere) AppendCondition(c types.Column, op types.Operation, val any) error {
	fn, ok := c.GetFilterFn(op)
	if !ok {
		return fmt.Errorf("no filter for %s", op)
	}
	sql, args, err := fn(context.Background(), val)
	if err != nil {
		return err
	}
	m.fragments = append_(m.fragments, sql)
	for _, a := range args {
		m.values = appendAny(m.values, a)
	}
	return nil
}

func append_(s []string, v string) []string { return append(s, v) }
func appendAny(s []any, v any) []any        { return append(s, v) }

type mockOrder struct {
	sqlpart.Order
	calls []string
}

func (m *mockOrder) OrderByColumn(c types.Column, d types.OrderDirection) {
	m.calls = append(m.calls, c.ToSQL(context.Background())+" "+string(d))
}
func (m *mockOrder) OrderBy(s string) { m.calls = append(m.calls, s) }

type mockGroup struct {
	sqlpart.Group
	cols []types.Column
}

func (m *mockGroup) GroupBy(cols ...types.Column) { m.cols = append(m.cols, cols...) }

type mockJoin struct {
	sqlpart.Join
	callbacks int
	bound     []string
	boundArgs [][]any
}

func (m *mockJoin) JOIN(func(ctx context.Context) string) { m.callbacks++ }
func (m *mockJoin) JOINOn(sql string, args ...any) {
	m.bound = append(m.bound, sql)
	m.boundArgs = append(m.boundArgs, args)
}

type mockLimitOffset struct {
	sqlpart.LimitOffset
	offset uint64
	limit  uint64
}

func (m *mockLimitOffset) SetOffset(v uint64) { m.offset = v }
func (m *mockLimitOffset) SetLimit(v uint64)  { m.limit = v }
func (m *mockLimitOffset) GetOffset() uint64  { return m.offset }
func (m *mockLimitOffset) GetLimit() uint64   { return m.limit }

// --- Applier — implements every <Op>Applier interface of the query package --

type mockApplier struct {
	ctx          context.Context
	storage      *mockColumnsStorage
	cols         *mockExecCols
	where        *mockWhere
	order        *mockOrder
	group        *mockGroup
	join         *mockJoin
	limit        *mockLimitOffset
	returningSet []types.Column // captured by SetReturning, asserted by tests
}

func newMockApplier() *mockApplier {
	cols := &mockExecCols{}
	storage := &mockColumnsStorage{
		byPtr:   map[any]types.Column{},
		execCol: cols,
	}
	return &mockApplier{
		ctx:     context.Background(),
		storage: storage,
		cols:    cols,
		where:   &mockWhere{},
		order:   &mockOrder{},
		group:   &mockGroup{},
		join:    &mockJoin{},
		limit:   &mockLimitOffset{},
	}
}

func (m *mockApplier) Ctx() context.Context                 { return m.ctx }
func (m *mockApplier) ColumnsStorage() types.ColumnsStorage { return m.storage }
func (m *mockApplier) Columns() types.ExecutionColumns      { return m.cols }
func (m *mockApplier) Where() sqlpart.Where                 { return m.where }
func (m *mockApplier) Order() sqlpart.Order                 { return m.order }
func (m *mockApplier) Group() sqlpart.Group                 { return m.group }
func (m *mockApplier) Join() sqlpart.Join                   { return m.join }
func (m *mockApplier) LimitOffset() sqlpart.LimitOffset     { return m.limit }
func (m *mockApplier) SetReturning(cols []types.Column)     { m.returningSet = cols }
