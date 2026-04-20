package linq

import (
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

func NewGroupBuilder(baseModel any) *GroupBuilder {
	return &GroupBuilder{
		model: baseModel,
	}
}

type GroupBuilder struct {
	model     any
	fieldPtrs []any
}

type GroupApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Group() sqlpart.Group
}

// selectableApplier marks a statement that exposes its SELECT column list
// (currently the read-style sqlselect: GetFirst, GetList). Used by the auto
// GROUP BY logic to discover non-aggregate columns when the user did not
// configure GroupBy explicitly.
type selectableApplier interface {
	Columns() types.ExecutionColumns
}

func (q *GroupBuilder) Apply(applier GroupApplier) error {
	group := applier.Group()
	if len(q.fieldPtrs) > 0 {
		storage := applier.ColumnsStorage()
		for _, fieldPtr := range q.fieldPtrs {
			col, err := storage.GetByFieldPtr(q.model, fieldPtr)
			if err != nil {
				return err
			}
			group.GroupBy(col)
		}
		return nil
	}
	// Auto GROUP BY: if any SELECT column is an aggregate and the user did not
	// configure GroupBy, group by every non-aggregate SELECT column. This keeps
	// the user-visible promise — "the type system covers what the SQL needs" —
	// without forcing them to mirror the SELECT list manually for every
	// aggregate virtual column they declare.
	selectable, ok := applier.(selectableApplier)
	if !ok {
		return nil
	}
	cols := selectable.Columns()
	if cols == nil {
		return nil
	}
	all := cols.GetAll()
	hasAggregate := false
	for _, c := range all {
		if c.IsAggregate() {
			hasAggregate = true
			break
		}
	}
	if !hasAggregate {
		return nil
	}
	for _, c := range all {
		if c.IsAggregate() {
			continue
		}
		group.GroupBy(c)
	}
	return nil
}

func (q *GroupBuilder) GroupBy(fieldsPtr ...any) {
	q.fieldPtrs = append(q.fieldPtrs, fieldsPtr...)
}
