package gerpo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/types"
	"github.com/insei/gerpo/virtual"
)

func TestName(t *testing.T) {
	b, _ := NewBuilder[test]()
	b.Table("test").
		Columns(func(m *test, columns *ColumnBuilder[test]) {
			columns.Column(&m.ID).WithInsertProtection().WithUpdateProtection()
			columns.Column(&m.CreatedAt)
			columns.Column(&m.UpdatedAt)
			columns.Column(&m.Name)
			columns.Column(&m.Age)
			columns.Column(&m.DeletedAt)
			columns.Virtual(&m.Bool).
				WithSQL(func(ctx context.Context) string {
					return `test.created_at > now()`
				}).
				WithBoolEqFilter(func(b *virtual.BoolEQFilterBuilder) {
					b.AddFalseSQLFn(func(ctx context.Context) string { return "test.created_at > now()" })
					b.AddTrueSQLFn(func(ctx context.Context) string { return "test.created_at < now()" })
				})
		}).
		BeforeInsert(func(ctx context.Context, m *test) {
			m.ID = 1
			m.CreatedAt = time.Now()
		}).
		BeforeUpdate(func(ctx context.Context, m *test) {
			updAt := time.Now()
			m.UpdatedAt = &updAt
		})

	repo, err := b.Build()
	repo.GetFirst(context.Background(), func(m *test, b *query.Helper[test]) {
		b.Select().Exclude(&m.Name)
		b.Where().
			Field(&m.Name).EQ("Ivan").
			OR().
			Group(func(t types.WhereTarget[test]) {
				t.
					Field(&m.Name).CT("any").
					AND().
					Field(&m.Age).GT(12)
			}).
			AND().
			Field(&m.Bool).EQ(true)
		b.RESTAPI().AppendFilters("id:neq:1||id:eq:2||id:eq:3")
		b.OrderBy().Field(&m.Name).DESC()
	})
	_ = err
	fmt.Println(repo, err)
}
