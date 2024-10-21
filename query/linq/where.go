package linq

import (
	"github.com/insei/gerpo/types"
)

type WhereBuilder struct {
	core *CoreBuilder
	opts []func(a types.ConditionBuilder)
}

func NewWhereBuilder(core *CoreBuilder) *WhereBuilder {
	return &WhereBuilder{
		core: core,
	}
}

func (q *WhereBuilder) Apply(condBuilder types.ConditionBuilder) {
	if len(q.opts) > 0 {
		condBuilder.StartGroup()
		for _, opt := range q.opts {
			opt(condBuilder)
		}
		condBuilder.EndGroup()
	}
}

func (q *WhereBuilder) IsEmpty() bool {
	return len(q.opts) == 0
}

func (q *WhereBuilder) Group(f func(t types.WhereTarget)) types.ANDOR {
	q.opts = append(q.opts, func(a types.ConditionBuilder) {
		a.StartGroup()
	})
	f(q)
	q.opts = append(q.opts, func(a types.ConditionBuilder) {
		a.EndGroup()
	})
	return q
}

type OperationFn func(operation types.Operation, val any) types.ANDOR

func (o OperationFn) EQ(val any) types.ANDOR {
	return o(types.OperationEQ, val)
}

func (o OperationFn) NEQ(val any) types.ANDOR {
	return o(types.OperationNEQ, val)
}

func (o OperationFn) CT(val any) types.ANDOR {
	return o(types.OperationCT, val)
}
func (o OperationFn) NCT(val any) types.ANDOR {
	return o(types.OperationNCT, val)
}

func (o OperationFn) GT(val any) types.ANDOR {
	return o(types.OperationGT, val)
}

func (o OperationFn) GTE(val any) types.ANDOR {
	return o(types.OperationGTE, val)
}

func (o OperationFn) LT(val any) types.ANDOR {
	return o(types.OperationLT, val)
}

func (o OperationFn) LTE(val any) types.ANDOR {
	return o(types.OperationLTE, val)
}

func (o OperationFn) BW(val any) types.ANDOR {
	return o(types.OperationBW, val)
}

func (o OperationFn) NBW(val any) types.ANDOR {
	return o(types.OperationNBW, val)
}

func (o OperationFn) EW(val any) types.ANDOR {
	return o(types.OperationEW, val)
}

func (o OperationFn) NEW(val any) types.ANDOR {
	return o(types.OperationNEW, val)
}
func (o OperationFn) IN(vals ...any) types.ANDOR {
	return o(types.OperationIN, vals)
}
func (o OperationFn) NIN(vals ...any) types.ANDOR {
	return o(types.OperationNIN, vals)
}
func (q *WhereBuilder) AND() types.WhereTarget {
	q.opts = append(q.opts, func(a types.ConditionBuilder) {
		a.AND()
	})
	return q
}

func (q *WhereBuilder) OR() types.WhereTarget {
	q.opts = append(q.opts, func(a types.ConditionBuilder) {
		a.OR()
	})
	return q
}

func (q *WhereBuilder) Field(fieldPtr any) types.WhereOperation {
	col := q.core.GetColumn(fieldPtr)
	return OperationFn(func(operation types.Operation, val any) types.ANDOR {
		q.opts = append(q.opts, func(a types.ConditionBuilder) {
			err := a.AppendCondition(col, operation, val)
			if err != nil {
				panic(err)
			}
		})
		return q
	})
}
