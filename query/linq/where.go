package linq

import (
	"fmt"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type WhereBuilder struct {
	model any
	ops   []whereOpEntry
}

type WhereApplier interface {
	Where() sqlpart.Where
	ColumnsStorage() types.ColumnsStorage
}

type whereOpKind uint8

const (
	opStartGroup whereOpKind = iota
	opEndGroup
	opAND
	opOR
	opFieldCondition
	opColumnCondition
)

// whereOpEntry stores one structural operation of a WHERE clause without using closures.
// Only a subset of fields is relevant per kind; see Apply for the dispatch.
type whereOpEntry struct {
	kind      whereOpKind
	column    types.Column
	fieldPtr  any
	operation types.Operation
	val       any
}

func NewWhereBuilder(baseModel any) *WhereBuilder {
	return &WhereBuilder{
		model: baseModel,
	}
}

func (q *WhereBuilder) Apply(applier WhereApplier) error {
	if len(q.ops) == 0 {
		return nil
	}
	w := applier.Where()
	w.StartGroup()
	for i := range q.ops {
		op := &q.ops[i]
		switch op.kind {
		case opStartGroup:
			w.StartGroup()
		case opEndGroup:
			w.EndGroup()
		case opAND:
			w.AND()
		case opOR:
			w.OR()
		case opFieldCondition:
			column, err := applier.ColumnsStorage().GetByFieldPtr(q.model, op.fieldPtr)
			if err != nil {
				return err
			}
			if err := w.AppendCondition(column, op.operation, op.val); err != nil {
				return err
			}
		case opColumnCondition:
			if op.column == nil {
				return fmt.Errorf("column is nil")
			}
			if err := w.AppendCondition(op.column, op.operation, op.val); err != nil {
				return err
			}
		}
	}
	w.EndGroup()
	return nil
}

func (q *WhereBuilder) IsEmpty() bool {
	return len(q.ops) == 0
}

func (q *WhereBuilder) Group(f func(t types.WhereTarget)) types.ANDOR {
	q.ops = append(q.ops, whereOpEntry{kind: opStartGroup})
	f(q)
	q.ops = append(q.ops, whereOpEntry{kind: opEndGroup})
	return q
}

// withIgnoreCase returns a case-insensitive variant of the provided operation if applicable, otherwise returns the original operation.
func withIgnoreCase(op types.Operation) types.Operation {
	switch op {
	case types.OperationContains, types.OperationNotContains,
		types.OperationStartsWith, types.OperationNotStartsWith,
		types.OperationEndsWith, types.OperationNotEndsWith:
		return op + "_ic"
	default:
		return op
	}
}

// whereOperation binds either a types.Column or a fieldPtr to its parent WhereBuilder
// so that the subsequent EQ/NEQ/… call can append a structured operation without a closure.
type whereOperation struct {
	parent   *WhereBuilder
	column   types.Column
	fieldPtr any
	isField  bool
}

func (o *whereOperation) push(operation types.Operation, val any) types.ANDOR {
	entry := whereOpEntry{
		operation: operation,
		val:       val,
	}
	if o.isField {
		entry.kind = opFieldCondition
		entry.fieldPtr = o.fieldPtr
	} else {
		entry.kind = opColumnCondition
		entry.column = o.column
	}
	o.parent.ops = append(o.parent.ops, entry)
	return o.parent
}

func (o *whereOperation) EQ(val any) types.ANDOR  { return o.push(types.OperationEQ, val) }
func (o *whereOperation) NEQ(val any) types.ANDOR { return o.push(types.OperationNEQ, val) }
func (o *whereOperation) GT(val any) types.ANDOR  { return o.push(types.OperationGT, val) }
func (o *whereOperation) GTE(val any) types.ANDOR { return o.push(types.OperationGTE, val) }
func (o *whereOperation) LT(val any) types.ANDOR  { return o.push(types.OperationLT, val) }
func (o *whereOperation) LTE(val any) types.ANDOR { return o.push(types.OperationLTE, val) }
func (o *whereOperation) IN(vals ...any) types.ANDOR {
	return o.push(types.OperationIN, vals)
}
func (o *whereOperation) NIN(vals ...any) types.ANDOR {
	return o.push(types.OperationNIN, vals)
}

func resolveIgnoreCase(base types.Operation, ignoreCase []bool) types.Operation {
	for _, v := range ignoreCase {
		if v {
			return withIgnoreCase(base)
		}
	}
	return base
}

func (o *whereOperation) StartsWith(val any, ignoreCase ...bool) types.ANDOR {
	return o.push(resolveIgnoreCase(types.OperationStartsWith, ignoreCase), val)
}
func (o *whereOperation) NotStartsWith(val any, ignoreCase ...bool) types.ANDOR {
	return o.push(resolveIgnoreCase(types.OperationNotStartsWith, ignoreCase), val)
}
func (o *whereOperation) EndsWith(val any, ignoreCase ...bool) types.ANDOR {
	return o.push(resolveIgnoreCase(types.OperationEndsWith, ignoreCase), val)
}
func (o *whereOperation) NotEndsWith(val any, ignoreCase ...bool) types.ANDOR {
	return o.push(resolveIgnoreCase(types.OperationNotEndsWith, ignoreCase), val)
}
func (o *whereOperation) Contains(val any, ignoreCase ...bool) types.ANDOR {
	return o.push(resolveIgnoreCase(types.OperationContains, ignoreCase), val)
}
func (o *whereOperation) NotContains(val any, ignoreCase ...bool) types.ANDOR {
	return o.push(resolveIgnoreCase(types.OperationNotContains, ignoreCase), val)
}
func (o *whereOperation) OP(operation types.Operation, val any) types.ANDOR {
	return o.push(operation, val)
}

func (q *WhereBuilder) AND() types.WhereTarget {
	q.ops = append(q.ops, whereOpEntry{kind: opAND})
	return q
}

func (q *WhereBuilder) OR() types.WhereTarget {
	q.ops = append(q.ops, whereOpEntry{kind: opOR})
	return q
}

func (q *WhereBuilder) Column(column types.Column) types.WhereOperation {
	return &whereOperation{parent: q, column: column}
}

func (q *WhereBuilder) Field(fieldPtr any) types.WhereOperation {
	return &whereOperation{parent: q, fieldPtr: fieldPtr, isField: true}
}
