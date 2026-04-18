package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insei/gerpo"
	extypes "github.com/insei/gerpo/executor/types"
	"github.com/insei/gerpo/query"
)

// BenchmarkGetFirstMock — временный бенчмарк без PostgreSQL для измерения heap allocations
// внутри gerpo. Не должен попасть в коммиты.
func BenchmarkGetFirstMock(b *testing.B) {
	type User struct {
		ID        uuid.UUID
		CreatedAt time.Time
		UpdatedAt *time.Time
		Name      string
		DeletedAt *time.Time
	}
	dateAt := time.Now().UTC()
	db := newMockDB()
	db.QueryContextFn = func(ctx context.Context, q string, args ...any) (extypes.Rows, error) {
		return &mockRows{max: 1}, nil
	}
	repo, err := gerpo.NewBuilder[User]().
		DB(db).
		Table("users").
		Columns(func(m *User, columns *gerpo.ColumnBuilder[User]) {
			columns.Field(&m.ID).AsColumn().WithUpdateProtection()
			columns.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
			columns.Field(&m.UpdatedAt).AsColumn().WithInsertProtection()
			columns.Field(&m.Name).AsColumn()
			columns.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
		}).
		WithSoftDeletion(func(m *User, softDeletion *gerpo.SoftDeletionBuilder[User]) {
			softDeletion.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				return &dateAt
			})
		}).
		WithQuery(func(m *User, h query.PersistentHelper[User]) {
			h.Where().Field(&m.DeletedAt).EQ(nil)
		}).
		Build()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetFirst(context.Background(), func(m *User, h query.GetFirstHelper[User]) {
			h.OrderBy().Field(&m.CreatedAt).ASC()
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}
