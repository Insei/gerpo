package sqlstmt

import (
	"context"
	"strings"
	"sync"

	"github.com/insei/gerpo/types"
)

type Count struct {
	*sqlselect

	table string
}

var countPool = sync.Pool{
	New: func() any {
		return &Count{sqlselect: newSelectEmpty()}
	},
}

func NewCount(ctx context.Context, table string, storage types.ColumnsStorage) *Count {
	f := countPool.Get().(*Count)
	f.table = table
	f.reset(ctx, storage)
	return f
}

// Release returns the statement to the pool. Must not be used after Release.
func (c *Count) Release() {
	c.table = ""
	c.columnsStorage = nil
	countPool.Put(c)
}

func (c *Count) SQL(_ ...Option) (string, []any, error) {
	if strings.TrimSpace(c.table) == "" {
		return "", nil, ErrTableIsNoSet
	}
	sb := strings.Builder{}
	sb.Grow(96)
	sb.WriteString("SELECT count(*) over() AS count FROM ")
	sb.WriteString(c.table)
	sb.WriteString(c.join.SQL())
	sb.WriteString(c.where.SQL())
	sb.WriteString(c.group.SQL())
	sb.WriteString(" LIMIT 1")
	return sb.String(), c.where.Values(), nil
}
