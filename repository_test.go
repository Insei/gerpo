package gerpo

import (
	"context"
	dbsql "database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/insei/gerpo/cache/ctx"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/virtual"
	//_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
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

var excludeRepositoryTest = true

func TestName(t *testing.T) {
	if excludeRepositoryTest {
		return
	}
	//db, err := otelsql.Open("postgres", "postgres://postgres:Admin@123@postgres.citmed:5432/test?sslmode=disable", otelsql.WithAttributes(
	//	semconv.DBSystemPostgreSQL,
	//))
	db, err := dbsql.Open("postgres", "postgres://postgres:Admin@123@postgres.citmed:5432/test?sslmode=disable")
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
		}).
		WithQuery(func(m *test, h query.PersistentUserHelper[test]) {
			h.Where().Field(&m.ID).LT(7)
			h.Exclude(&m.UpdatedAt, &m.ID)
			h.LeftJoin(func(ctx context.Context) string {
				return ``
			})
		})
	repo, err := b.Build()

	ctxCache := ctx.NewCtxCache(context.Background())
	_ = []int{1, 2, 3, 4, 5, 6}
	list, err := repo.GetList(ctxCache, func(m *test, h query.GetListUserHelper[test]) {
		h.Where().Field(&m.ID).IN(1, 2, 3, 4, 5, 6)
	})

	start := time.Now()
	model, err := repo.GetFirst(ctxCache)
	elapsed := time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("FIRST Repo db get one %s", elapsed)

	start = time.Now()
	model, err = repo.GetFirst(ctxCache, func(m *test, b query.GetFirstUserHelper[test]) {
		b.Where().
			Field(&m.ID).EQ(2)
		b.OrderBy().Field(&m.Name).DESC()
	})
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("SECOND Repo db get one %s", elapsed)

	start = time.Now()
	model, err = repo.GetFirst(ctxCache, func(m *test, b query.GetFirstUserHelper[test]) {
		b.Where().
			Field(&m.ID).EQ(2)
		b.OrderBy().Field(&m.Name).DESC()
	})
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo get same one from cache %s", elapsed)

	start = time.Now()
	count, err := repo.Count(ctxCache, func(m *test, h query.CountUserHelper[test]) {
		h.Where().
			Field(&m.ID).EQ(1)
	})
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo db count %s", elapsed)

	start = time.Now()
	count, err = repo.Count(ctxCache, func(m *test, h query.CountUserHelper[test]) {
		h.Where().
			Field(&m.ID).EQ(1)
	})
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo count same from cache %s", elapsed)

	start = time.Now()
	count, err = repo.Count(ctxCache)
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo count ALL from db %s", elapsed)

	start = time.Now()
	count, err = repo.Count(ctxCache)
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo count ALL from cache %s", elapsed)

	_ = err
	err = repo.Insert(ctxCache, &test{
		Name: "Firts",
		Age:  330,
		Bool: true,
	}, func(m *test, h query.InsertUserHelper[test]) {
		h.Exclude()
	})
	if err != nil {
		log.Print(err)
	}
	start = time.Now()
	list, err = repo.GetList(ctxCache) //ALL
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo GetList ALL from db %s", elapsed)

	start = time.Now()
	list, err = repo.GetList(ctxCache) //ALL
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo GetList ALL from cache %s", elapsed)

	start = time.Now()
	list, err = repo.GetList(ctxCache, func(m *test, h query.GetListUserHelper[test]) {
		h.Page(1).Size(2)
		h.OrderBy().Field(&m.CreatedAt).DESC()
	})
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo GetList limit 2, page 1 from db %s", elapsed)

	start = time.Now()
	list, err = repo.GetList(ctxCache, func(m *test, h query.GetListUserHelper[test]) {
		h.Page(1).Size(2)
		h.OrderBy().Field(&m.CreatedAt).DESC()
	})
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo GetList limit 2, page 1 from cache %s", elapsed)

	model.Age = 777
	err = repo.Update(ctxCache, model, func(m *test, h query.UpdateUserHelper[test]) {
		h.Where().Field(&m.ID).EQ(5)
	})
	if err != nil {
		log.Print(err)
	}

	deletedCount, err := repo.Delete(ctxCache, func(m *test, h query.DeleteUserHelper[test]) {
		h.Where().Field(&m.ID).EQ(1)
	})
	if err != nil {
		log.Print(err)
	}
	fmt.Println(list, model, count, deletedCount)
}
