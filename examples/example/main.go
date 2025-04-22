package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/insei/gerpo/executor/adapters/databasesql"
	"github.com/insei/gerpo/executor/adapters/placeholder"
	_ "github.com/lib/pq"

	"github.com/insei/gerpo"
	"github.com/insei/gerpo/executor/cache/ctx"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/virtual"
)

type test struct {
	ID          int
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	Name        string
	Age         int
	Bool        bool
	DeletedAt   *time.Time
	DeletedTest bool
}

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:Admin@123@localhost:5432/test?sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	dbWrap := databasesql.NewAdapter(db, databasesql.WithPlaceholder(placeholder.Dollar))
	b := gerpo.NewBuilder[test]().
		DB(dbWrap).
		Table("tests").
		Columns(func(m *test, columns *gerpo.ColumnBuilder[test]) {
			columns.Field(&m.ID).AsColumn().WithInsertProtection().WithUpdateProtection()
			columns.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
			columns.Field(&m.UpdatedAt).AsColumn().WithInsertProtection()
			columns.Field(&m.Name).AsColumn()
			columns.Field(&m.Age).AsColumn()
			columns.Field(m.DeletedAt).AsColumn().WithUpdateProtection().WithInsertProtection()

			columns.Field(&m.Bool).AsVirtual().
				WithSQL(func(ctx context.Context) string {
					return `tests.created_at > now()`
				}).
				WithBoolEqFilter(func(b *virtual.BoolEQFilterBuilder) {
					b.AddFalseSQLFn(func(ctx context.Context) string { return "tests.created_at > now()" })
					b.AddTrueSQLFn(func(ctx context.Context) string { return "tests.created_at < now()" })
				})
		}).
		WithBeforeInsert(func(ctx context.Context, m *test) {
			m.CreatedAt = time.Now()
		}).
		WithBeforeUpdate(func(ctx context.Context, m *test) {
			updAt := time.Now()
			m.UpdatedAt = &updAt
		}).
		WithQuery(func(m *test, h query.PersistentHelper[test]) {
			h.Where().Field(&m.ID).LT(7)
			h.Exclude(m.UpdatedAt, m.ID)
			h.LeftJoin(func(ctx context.Context) string {
				return ``
			})
		})
	repo, err := b.Build()

	ctxCache := ctx.NewCtxCache(context.Background())
	_ = []int{1, 2, 3, 4, 5, 6}
	list, err := repo.GetList(ctxCache, func(m *test, h query.GetListHelper[test]) {
		h.Where().Field(&m.ID).IN(1, 2, 3, 4, 5, 6)
		h.Exclude(&m.UpdatedAt, &m.ID)
	})

	tx, err := dbWrap.BeginTx(ctxCache)
	if err != nil {
		log.Fatal(err)
	}
	repoTx, err := repo.Tx(tx)
	if err != nil {
		log.Fatal(err)
	}
	start := time.Now()
	model, err := repoTx.GetFirst(ctxCache)
	elapsed := time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("FIRST Repo db get one %s", elapsed)

	start = time.Now()
	model, err = repoTx.GetFirst(ctxCache, func(m *test, b query.GetFirstHelper[test]) {
		b.Where().
			Field(&m.ID).EQ(2)
		b.OrderBy().Field(m.Name).DESC()
		b.Exclude(&m.UpdatedAt, m.ID)
	})
	err = repoTx.Insert(ctxCache, model)
	if err != nil {
		log.Print(err)
	}
	err = tx.Rollback()
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("SECOND Repo db get one %s", elapsed)

	start = time.Now()
	model, err = repo.GetFirst(ctxCache, func(m *test, b query.GetFirstHelper[test]) {
		b.Where().
			Field(m.ID).EQ(2)
		b.OrderBy().Field(&m.Name).DESC()
	})
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo get same one from cache %s", elapsed)

	start = time.Now()
	count, err := repo.Count(ctxCache, func(m *test, h query.CountHelper[test]) {
		h.Where().
			Field(&m.ID).EQ(1)
	})
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo db count %s", elapsed)

	start = time.Now()
	count, err = repo.Count(ctxCache, func(m *test, h query.CountHelper[test]) {
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
	}, func(m *test, h query.InsertHelper[test]) {
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
	list, err = repo.GetList(ctxCache, func(m *test, h query.GetListHelper[test]) {
		h.Page(1).Size(2)
		h.OrderBy().Field(&m.CreatedAt).DESC()
	})
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo GetList limit 2, page 1 from db %s", elapsed)

	start = time.Now()
	list, err = repo.GetList(ctxCache, func(m *test, h query.GetListHelper[test]) {
		h.Page(1).Size(2)
		h.OrderBy().Field(&m.CreatedAt).DESC()
	})
	elapsed = time.Since(start)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Repo GetList limit 2, page 1 from cache %s", elapsed)

	model.Age = 777
	_, err = repo.Update(ctxCache, model, func(m *test, h query.UpdateHelper[test]) {
		h.Where().Field(&m.ID).EQ(5)
	})
	if err != nil {
		log.Print(err)
	}

	deletedCount, err := repo.Delete(ctxCache, func(m *test, h query.DeleteHelper[test]) {
		h.Where().Field(&m.ID).EQ(1)
	})
	if err != nil {
		log.Print(err)
	}
	fmt.Println(list, model, count, deletedCount)
}
