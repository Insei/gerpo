package gerpo

import (
	"context"
	dbsql "database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/insei/gerpo/cache"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/virtual"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//type test struct {
//	ID        int        `json:"id"`
//	CreatedAt time.Time  `json:"created_at"`
//	UpdatedAt *time.Time `json:"updated_at"`
//	Name      string     `json:"name"`
//	Age       int        `json:"age"`
//	Bool      bool       `json:"bool"`
//	DeletedAt *time.Time `json:"deleted_at"`
//}

func deferTest() (test int, err error) {
	defer func() {
		fmt.Println(test, err)
	}()
	return 22, fmt.Errorf("Test")
}

func TestName(t *testing.T) {
	deferTest()

	db, err := dbsql.Open("pgx", "postgres://postgres:Admin@123@postgres.citmed:5432/test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	b := NewBuilder[test]().
		DB(db).
		Table("tests").
		Columns(func(m *test, columns *ColumnBuilder[test]) {
			columns.Column(&m.ID).WithInsertProtection().WithUpdateProtection()
			columns.Column(&m.CreatedAt).WithUpdateProtection()
			columns.Column(&m.UpdatedAt).WithInsertProtection()
			columns.Column(&m.Name)
			columns.Column(&m.Age)
			columns.Column(&m.DeletedAt).WithUpdateProtection().WithInsertProtection()
			columns.Virtual(&m.Bool).
				WithSQL(func(ctx context.Context) string {
					return `tests.created_at > now()`
				}).
				WithBoolEqFilter(func(b *virtual.BoolEQFilterBuilder) {
					b.AddFalseSQLFn(func(ctx context.Context) string { return "tests.created_at > now()" })
					b.AddTrueSQLFn(func(ctx context.Context) string { return "tests.created_at < now()" })
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

	ctx := cache.NewCtxCache(context.Background())

	start := time.Now()
	model, err := repo.GetFirst(ctx)
	elapsed := time.Since(start)
	log.Printf("FIRST Repo db get one %s", elapsed)

	start = time.Now()
	model, err = repo.GetFirst(ctx, func(m *test, b query.GetFirstUserHelper[test]) {
		b.Where().
			Field(&m.ID).EQ(2)
		b.OrderBy().Field(&m.Name).DESC()
	})
	elapsed = time.Since(start)
	log.Printf("SECOND Repo db get one %s", elapsed)

	start = time.Now()
	model, err = repo.GetFirst(ctx, func(m *test, b query.GetFirstUserHelper[test]) {
		b.Where().
			Field(&m.ID).EQ(2)
		b.OrderBy().Field(&m.Name).DESC()
	})
	elapsed = time.Since(start)
	log.Printf("Repo get same one from cache %s", elapsed)

	start = time.Now()
	count, err := repo.Count(ctx, func(m *test, h query.CountUserHelper[test]) {
		h.Where().
			Field(&m.ID).EQ(1)
	})
	elapsed = time.Since(start)
	log.Printf("Repo db count %s", elapsed)

	start = time.Now()
	count, err = repo.Count(ctx, func(m *test, h query.CountUserHelper[test]) {
		h.Where().
			Field(&m.ID).EQ(1)
	})
	elapsed = time.Since(start)
	log.Printf("Repo count same from cache %s", elapsed)

	start = time.Now()
	count, err = repo.Count(ctx)
	elapsed = time.Since(start)
	log.Printf("Repo count ALL from db %s", elapsed)

	start = time.Now()
	count, err = repo.Count(ctx)
	elapsed = time.Since(start)
	log.Printf("Repo count ALL from cache %s", elapsed)

	_ = err
	err = repo.Insert(ctx, &test{
		Name: "Firts",
		Age:  330,
		Bool: true,
	}, func(m *test, h query.InsertUserHelper[test]) {
		h.Exclude()
	})

	start = time.Now()
	list, err := repo.GetList(ctx) //ALL
	elapsed = time.Since(start)
	log.Printf("Repo GetList ALL from db %s", elapsed)

	start = time.Now()
	list, err = repo.GetList(ctx) //ALL
	elapsed = time.Since(start)
	log.Printf("Repo GetList ALL from cache %s", elapsed)

	start = time.Now()
	list, err = repo.GetList(ctx, func(m *test, h query.GetListUserHelper[test]) {
		h.Page(1).Size(2)
		h.OrderBy().Field(&m.CreatedAt).DESC()
	})
	elapsed = time.Since(start)
	log.Printf("Repo GetList limit 2, page 1 from db %s", elapsed)

	start = time.Now()
	list, err = repo.GetList(ctx, func(m *test, h query.GetListUserHelper[test]) {
		h.Page(1).Size(2)
		h.OrderBy().Field(&m.CreatedAt).DESC()
	})
	elapsed = time.Since(start)
	log.Printf("Repo GetList limit 2, page 1 from cache %s", elapsed)

	model.Age = 555
	err = repo.Update(ctx, model, func(m *test, h query.UpdateUserHelper[test]) {
		h.Where().Field(&m.ID).EQ(5)
	})

	deletedCount, err := repo.Delete(ctx, func(m *test, h query.DeleteUserHelper[test]) {
		h.Where().Field(&m.ID).EQ(1)
	})
	fmt.Println(list, model, count, deletedCount)
}
