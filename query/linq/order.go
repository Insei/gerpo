package linq

import (
	"fmt"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

func NewOrderBuilder(baseModel any) *OrderBuilder {
	return &OrderBuilder{
		model: baseModel,
	}
}

type OrderBuilder struct {
	model any
	ops   []orderOpEntry
}

type OrderApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Order() sqlpart.Order
}

type orderOpKind uint8

const (
	orderKindColumn orderOpKind = iota
	orderKindField
)

type orderOpEntry struct {
	kind      orderOpKind
	column    types.Column
	fieldPtr  any
	direction types.OrderDirection
}

func (q *OrderBuilder) Apply(applier OrderApplier) error {
	if len(q.ops) == 0 {
		return nil
	}
	o := applier.Order()
	for i := range q.ops {
		op := &q.ops[i]
		switch op.kind {
		case orderKindColumn:
			if op.column == nil {
				return fmt.Errorf("column is nil")
			}
			o.OrderByColumn(op.column, op.direction)
		case orderKindField:
			column, err := applier.ColumnsStorage().GetByFieldPtr(q.model, op.fieldPtr)
			if err != nil {
				return err
			}
			o.OrderByColumn(column, op.direction)
		}
	}
	return nil
}

// orderDirection binds either a column or a field pointer to its parent OrderBuilder
// so ASC/DESC can append a structured operation without a closure.
type orderDirection struct {
	parent   *OrderBuilder
	column   types.Column
	fieldPtr any
	isField  bool
}

func (d *orderDirection) push(direction types.OrderDirection) *OrderBuilder {
	entry := orderOpEntry{direction: direction}
	if d.isField {
		entry.kind = orderKindField
		entry.fieldPtr = d.fieldPtr
	} else {
		entry.kind = orderKindColumn
		entry.column = d.column
	}
	d.parent.ops = append(d.parent.ops, entry)
	return d.parent
}

func (d *orderDirection) ASC() types.OrderTarget  { return d.push(types.OrderDirectionASC) }
func (d *orderDirection) DESC() types.OrderTarget { return d.push(types.OrderDirectionDESC) }

func (q *OrderBuilder) Column(column types.Column) types.OrderOperation {
	return &orderDirection{parent: q, column: column}
}

func (q *OrderBuilder) Field(fieldPtr any) types.OrderOperation {
	return &orderDirection{parent: q, fieldPtr: fieldPtr, isField: true}
}
