package query

import (
	"github.com/insei/gerpo/types"
)

type WhereBuilderFactory[TModel any] struct {
	model   *TModel
	columns *types.ColumnsStorage
}

func NewWhereBuilderFabric[TModel any](model *TModel, columns *types.ColumnsStorage) *WhereBuilderFactory[TModel] {
	return &WhereBuilderFactory[TModel]{
		model:   model,
		columns: columns,
	}
}

func (f *WhereBuilderFactory[TModel]) New() *WhereBuilder[TModel] {
	return &WhereBuilder[TModel]{
		fabric: f,
		opts:   nil,
	}
}

type WhereBuilder[TModel any] struct {
	fabric *WhereBuilderFactory[TModel]
	opts   []func(b *StringSQLWhereBuilder)
}

func (q *WhereBuilder[TModel]) Apply(strSQLBuilder *StringSQLWhereBuilder) {
	for _, opt := range q.opts {
		opt(strSQLBuilder)
	}
}

func (q *WhereBuilder[TModel]) Group(f func(t types.WhereTarget[TModel])) types.ANDOR[TModel] {
	q.opts = append(q.opts, func(b *StringSQLWhereBuilder) {
		b.StartGroup()
	})
	f(q)
	q.opts = append(q.opts, func(b *StringSQLWhereBuilder) {
		b.EndGroup()
	})
	return q
}

type OperationFn[TModel any] func(operation types.Operation, val any) types.ANDOR[TModel]

func (o OperationFn[TModel]) EQ(val any) types.ANDOR[TModel] {
	return o(types.OperationEQ, val)
}

func (o OperationFn[TModel]) NEQ(val any) types.ANDOR[TModel] {
	return o(types.OperationNEQ, val)
}

func (o OperationFn[TModel]) CT(val any) types.ANDOR[TModel] {
	return o(types.OperationCT, val)
}
func (o OperationFn[TModel]) NCT(val any) types.ANDOR[TModel] {
	return o(types.OperationNCT, val)
}

func (o OperationFn[TModel]) GT(val any) types.ANDOR[TModel] {
	return o(types.OperationGT, val)
}

func (o OperationFn[TModel]) GTE(val any) types.ANDOR[TModel] {
	return o(types.OperationGTE, val)
}

func (o OperationFn[TModel]) LT(val any) types.ANDOR[TModel] {
	return o(types.OperationLT, val)
}

func (o OperationFn[TModel]) LTE(val any) types.ANDOR[TModel] {
	return o(types.OperationLTE, val)
}

func (o OperationFn[TModel]) BW(val any) types.ANDOR[TModel] {
	return o(types.OperationBW, val)
}

func (o OperationFn[TModel]) NBW(val any) types.ANDOR[TModel] {
	return o(types.OperationNBW, val)
}

func (o OperationFn[TModel]) EW(val any) types.ANDOR[TModel] {
	return o(types.OperationEW, val)
}

func (o OperationFn[TModel]) NEW(val any) types.ANDOR[TModel] {
	return o(types.OperationNEW, val)
}

func (q *WhereBuilder[TModel]) AND() types.WhereTarget[TModel] {
	q.opts = append(q.opts, func(b *StringSQLWhereBuilder) {
		b.AND()
	})
	return q
}

func (q *WhereBuilder[TModel]) OR() types.WhereTarget[TModel] {
	q.opts = append(q.opts, func(b *StringSQLWhereBuilder) {
		b.OR()
	})
	return q
}

func (q *WhereBuilder[TModel]) Field(fieldPtr any) types.WhereOperation[TModel] {
	col, err := q.fabric.columns.GetByFieldPtr(q.fabric.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	return OperationFn[TModel](func(operation types.Operation, val any) types.ANDOR[TModel] {
		q.opts = append(q.opts, func(b *StringSQLWhereBuilder) {
			err := b.AppendCondition(col, operation, val)
			if err != nil {
				panic(err)
			}
		})
		return q
	})
}
