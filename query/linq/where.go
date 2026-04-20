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

// whereOperation binds either a types.Column or a fieldPtr to its parent WhereBuilder
// so that the subsequent EQ/NotEQ/… call can append a structured operation without a closure.
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

// Universal operators.

func (o *whereOperation) EQ(val any) types.ANDOR    { return o.push(types.OperationEQ, val) }
func (o *whereOperation) NotEQ(val any) types.ANDOR { return o.push(types.OperationNotEQ, val) }
func (o *whereOperation) GT(val any) types.ANDOR    { return o.push(types.OperationGT, val) }
func (o *whereOperation) GTE(val any) types.ANDOR   { return o.push(types.OperationGTE, val) }
func (o *whereOperation) LT(val any) types.ANDOR    { return o.push(types.OperationLT, val) }
func (o *whereOperation) LTE(val any) types.ANDOR   { return o.push(types.OperationLTE, val) }

func (o *whereOperation) In(vals ...any) types.ANDOR    { return o.push(types.OperationIn, vals) }
func (o *whereOperation) NotIn(vals ...any) types.ANDOR { return o.push(types.OperationNotIn, vals) }

// String operators, case-sensitive.

func (o *whereOperation) Contains(val any) types.ANDOR {
	return o.push(types.OperationContains, val)
}
func (o *whereOperation) NotContains(val any) types.ANDOR {
	return o.push(types.OperationNotContains, val)
}
func (o *whereOperation) StartsWith(val any) types.ANDOR {
	return o.push(types.OperationStartsWith, val)
}
func (o *whereOperation) NotStartsWith(val any) types.ANDOR {
	return o.push(types.OperationNotStartsWith, val)
}
func (o *whereOperation) EndsWith(val any) types.ANDOR {
	return o.push(types.OperationEndsWith, val)
}
func (o *whereOperation) NotEndsWith(val any) types.ANDOR {
	return o.push(types.OperationNotEndsWith, val)
}

// Case-insensitive "Fold" variants — mirrors strings.EqualFold naming.

func (o *whereOperation) EQFold(val any) types.ANDOR {
	return o.push(types.OperationEQFold, val)
}
func (o *whereOperation) NotEQFold(val any) types.ANDOR {
	return o.push(types.OperationNotEQFold, val)
}
func (o *whereOperation) ContainsFold(val any) types.ANDOR {
	return o.push(types.OperationContainsFold, val)
}
func (o *whereOperation) NotContainsFold(val any) types.ANDOR {
	return o.push(types.OperationNotContainsFold, val)
}
func (o *whereOperation) StartsWithFold(val any) types.ANDOR {
	return o.push(types.OperationStartsWithFold, val)
}
func (o *whereOperation) NotStartsWithFold(val any) types.ANDOR {
	return o.push(types.OperationNotStartsWithFold, val)
}
func (o *whereOperation) EndsWithFold(val any) types.ANDOR {
	return o.push(types.OperationEndsWithFold, val)
}
func (o *whereOperation) NotEndsWithFold(val any) types.ANDOR {
	return o.push(types.OperationNotEndsWithFold, val)
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
