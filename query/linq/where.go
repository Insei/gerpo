package linq

import (
	"fmt"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type WhereBuilder struct {
	model any
	opts  []func(applier WhereApplier) error
}

type WhereApplier interface {
	Where() sqlpart.Where
	ColumnsStorage() types.ColumnsStorage
}

func NewWhereBuilder(baseModel any) *WhereBuilder {
	return &WhereBuilder{
		model: baseModel,
	}
}

func (q *WhereBuilder) Apply(applier WhereApplier) error {
	if len(q.opts) > 0 {
		applier.Where().StartGroup()
		for _, opt := range q.opts {
			err := opt(applier)
			if err != nil {
				return err
			}
		}
		applier.Where().EndGroup()
	}
	return nil
}

func (q *WhereBuilder) IsEmpty() bool {
	return len(q.opts) == 0
}

func (q *WhereBuilder) Group(f func(t types.WhereTarget)) types.ANDOR {
	q.opts = append(q.opts, func(applier WhereApplier) error {
		applier.Where().StartGroup()
		return nil
	})
	f(q)
	q.opts = append(q.opts, func(applier WhereApplier) error {
		applier.Where().EndGroup()
		return nil
	})
	return q
}

// withIgnoreCase returns a case-insensitive variant of the provided operation if applicable, otherwise returns the original operation.
func withIgnoreCase(op types.Operation) types.Operation {
	switch op {
	case types.OperationCT, types.OperationNCT, types.OperationEW, types.OperationNEW, types.OperationBW, types.OperationNBW:
		return op + "_ic"
	default:
		return op
	}
}

type OperationFn func(operation types.Operation, val any) types.ANDOR

func (o OperationFn) EQ(val any) types.ANDOR {
	return o(types.OperationEQ, val)
}
func (o OperationFn) NEQ(val any) types.ANDOR {
	return o(types.OperationNEQ, val)
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
func (o OperationFn) IN(vals ...any) types.ANDOR {
	return o(types.OperationIN, vals)
}
func (o OperationFn) NIN(vals ...any) types.ANDOR {
	return o(types.OperationNIN, vals)
}
func (o OperationFn) BW(val any, ignoreCase ...bool) types.ANDOR {
	op := types.OperationBW
	for _, opt := range ignoreCase {
		if opt {
			op = withIgnoreCase(op)
		}
	}
	return o(op, val)
}
func (o OperationFn) NBW(val any, ignoreCase ...bool) types.ANDOR {
	op := types.OperationNBW
	for _, opt := range ignoreCase {
		if opt {
			op = withIgnoreCase(op)
		}
	}
	return o(op, val)
}
func (o OperationFn) EW(val any, ignoreCase ...bool) types.ANDOR {
	op := types.OperationEW
	for _, opt := range ignoreCase {
		if opt {
			op = withIgnoreCase(op)
		}
	}
	return o(op, val)
}
func (o OperationFn) NEW(val any, ignoreCase ...bool) types.ANDOR {
	op := types.OperationNEW
	for _, opt := range ignoreCase {
		if opt {
			op = withIgnoreCase(op)
		}
	}
	return o(op, val)
}
func (o OperationFn) CT(val any, ignoreCase ...bool) types.ANDOR {
	op := types.OperationCT
	for _, opt := range ignoreCase {
		if opt {
			op = withIgnoreCase(op)
		}
	}
	return o(op, val)
}
func (o OperationFn) NCT(val any, ignoreCase ...bool) types.ANDOR {
	op := types.OperationNCT
	for _, opt := range ignoreCase {
		if opt {
			op = withIgnoreCase(op)
		}
	}
	return o(op, val)
}
func (o OperationFn) OP(operation types.Operation, val any) types.ANDOR {
	return o(operation, val)
}

func (q *WhereBuilder) AND() types.WhereTarget {
	q.opts = append(q.opts, func(applier WhereApplier) error {
		applier.Where().AND()
		return nil
	})
	return q
}

func (q *WhereBuilder) OR() types.WhereTarget {
	q.opts = append(q.opts, func(applier WhereApplier) error {
		applier.Where().OR()
		return nil
	})
	return q
}

func (q *WhereBuilder) Column(column types.Column) types.WhereOperation {
	return OperationFn(func(operation types.Operation, val any) types.ANDOR {
		q.opts = append(q.opts, func(applier WhereApplier) error {
			if column == nil {
				return fmt.Errorf("column is nil")
			}
			err := applier.Where().AppendCondition(column, operation, val)
			if err != nil {
				return err
			}
			return err
		})
		return q
	})
}

func (q *WhereBuilder) Field(fieldPtr any) types.WhereOperation {
	return OperationFn(func(operation types.Operation, val any) types.ANDOR {
		q.opts = append(q.opts, func(applier WhereApplier) error {
			column, err := applier.ColumnsStorage().GetByFieldPtr(q.model, fieldPtr)
			if err != nil {
				return err
			}
			err = applier.Where().AppendCondition(column, operation, val)
			if err != nil {
				return err
			}
			return nil
		})
		return q
	})
}
