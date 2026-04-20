package gerpo_test

// The examples below appear on pkg.go.dev next to the corresponding methods
// and types. They do not run during `go test ./...` (no // Output: line) —
// each one would require a live database — but the compiler still checks them,
// so the snippets cannot rot away from the public API.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/insei/gerpo"
	"github.com/insei/gerpo/executor"
	"github.com/insei/gerpo/executor/adapters/pgx5"
	cachectx "github.com/insei/gerpo/executor/cache/ctx"
	"github.com/insei/gerpo/query"
)

// User is the working example throughout these snippets — a typical entity
// with a UUID primary key, a nullable email and timestamp tracking columns.
type User struct {
	ID        uuid.UUID
	Name      string
	Email     *string
	Age       int
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

// exampleRepo returns a placeholder repository so the snippet methods on
// pkg.go.dev focus on the API at the call site instead of the boilerplate
// that produces the repo. Replace with the real repository in your code.
func exampleRepo() gerpo.Repository[User] { return nil }

// exampleAdapter returns a placeholder adapter for the same reason.
func exampleAdapter() executor.Adapter { return nil }

// ExampleNew shows the minimum chain to assemble a typed repository
// against a pgx v5 pool.
func ExampleNew() {
	pool, err := pgxpool.New(context.Background(), "postgres://localhost/db")
	if err != nil {
		panic(err)
	}

	repo, err := gerpo.New[User]().
		DB(pgx5.NewPoolAdapter(pool)).
		Table("users").
		Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
			c.Field(&m.ID).OmitOnUpdate()
			c.Field(&m.Name)
			c.Field(&m.Email)
			c.Field(&m.Age)
			c.Field(&m.CreatedAt).OmitOnUpdate()
			c.Field(&m.UpdatedAt).OmitOnInsert()
		}).
		Build()
	if err != nil {
		panic(err)
	}
	_ = repo
}

// ExampleRepository_GetFirst fetches a single record by primary key and
// translates the absence of a row into the domain layer.
func ExampleRepository_GetFirst() {
	repo := exampleRepo()

	u, err := repo.GetFirst(context.Background(), func(m *User, h query.GetFirstHelper[User]) {
		h.Where().Field(&m.Email).EQ("alice@example.com")
		h.OrderBy().Field(&m.CreatedAt).DESC()
	})
	switch {
	case errors.Is(err, gerpo.ErrNotFound):
		fmt.Println("user not found")
	case err != nil:
		panic(err)
	default:
		fmt.Println(u.Name)
	}
}

// ExampleRepository_GetList shows filtering, ordering and pagination in one
// call. Adult users are returned 20 per page, newest first.
func ExampleRepository_GetList() {
	repo := exampleRepo()

	users, err := repo.GetList(context.Background(), func(m *User, h query.GetListHelper[User]) {
		h.Where().Field(&m.Age).GTE(18)
		h.OrderBy().Field(&m.CreatedAt).DESC()
		h.Page(1).Size(20)
	})
	if err != nil {
		panic(err)
	}
	for _, u := range users {
		fmt.Println(u.Name)
	}
}

// ExampleRepository_Insert demonstrates Insert together with InsertHelper to
// let the database default created_at instead of a Go-side timestamp.
func ExampleRepository_Insert() {
	repo := exampleRepo()

	u := &User{
		ID:   uuid.New(),
		Name: "Bob",
		Age:  30,
	}
	if err := repo.Insert(context.Background(), u, func(m *User, h query.InsertHelper[User]) {
		h.Exclude(&m.CreatedAt) // let the DB DEFAULT NOW()
	}); err != nil {
		panic(err)
	}
}

// ExampleRepository_Update updates a single column with Only and reports the
// number of rows touched.
func ExampleRepository_Update() {
	repo := exampleRepo()
	userID := uuid.New()

	u := &User{ID: userID, Name: "Bob the Builder"}
	rows, err := repo.Update(context.Background(), u, func(m *User, h query.UpdateHelper[User]) {
		h.Where().Field(&m.ID).EQ(userID)
		h.Only(&m.Name) // SET name = ?, leave everything else alone
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("updated %d row(s)\n", rows)
}

// ExampleRepository_Delete removes records that match a WHERE clause.
func ExampleRepository_Delete() {
	repo := exampleRepo()
	userID := uuid.New()

	rows, err := repo.Delete(context.Background(), func(m *User, h query.DeleteHelper[User]) {
		h.Where().Field(&m.ID).EQ(userID)
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("deleted %d row(s)\n", rows)
}

// ExampleWithSoftDeletion swaps a physical DELETE for an UPDATE of a marker
// column. The persistent WHERE in WithQuery hides soft-deleted rows from any
// future SELECT or COUNT.
func ExampleWithSoftDeletion() {
	pool, _ := pgxpool.New(context.Background(), "postgres://localhost/db")

	repo, err := gerpo.New[User]().
		DB(pgx5.NewPoolAdapter(pool)).
		Table("users").
		Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
			c.Field(&m.ID).OmitOnUpdate()
			c.Field(&m.Name)
			c.Field(&m.DeletedAt).OmitOnInsert()
		}).
		WithQuery(func(m *User, h query.PersistentHelper[User]) {
			h.Where().Field(&m.DeletedAt).EQ(nil) // hide soft-deleted rows
		}).
		WithSoftDeletion(func(m *User, b *gerpo.SoftDeletionBuilder[User]) {
			b.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				now := time.Now().UTC()
				return &now
			})
		}).
		Build()
	if err != nil {
		panic(err)
	}
	_ = repo
}

// ExampleWithErrorTransformer maps gerpo's sentinel errors to a domain error
// so the upper layers do not need to import the gerpo package.
func ExampleWithErrorTransformer() {
	var ErrUserNotFound = errors.New("user: not found")

	pool, _ := pgxpool.New(context.Background(), "postgres://localhost/db")

	repo, err := gerpo.New[User]().
		DB(pgx5.NewPoolAdapter(pool)).
		Table("users").
		Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
			c.Field(&m.ID).OmitOnUpdate()
			c.Field(&m.Name)
		}).
		WithErrorTransformer(func(err error) error {
			if errors.Is(err, gerpo.ErrNotFound) {
				return ErrUserNotFound
			}
			return err
		}).
		Build()
	if err != nil {
		panic(err)
	}
	_ = repo
}

// ExampleRepository_Tx shows the standard transactional pattern: open a
// driver transaction, wrap the repository with .Tx(tx), then commit or roll
// back. RollbackUnlessCommitted is safe to defer even after a successful Commit.
func ExampleRepository_Tx() {
	adapter := exampleAdapter()
	repo := exampleRepo()

	ctx := context.Background()
	tx, err := adapter.BeginTx(ctx)
	if err != nil {
		panic(err)
	}
	defer func() { _ = tx.RollbackUnlessCommitted() }()

	txRepo := repo.Tx(tx)
	if err := txRepo.Insert(ctx, &User{ID: uuid.New(), Name: "Carol"}); err != nil {
		return
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

// ExampleWithTracer wires gerpo into an OpenTelemetry-style tracer without
// pulling the OTel package into the gerpo dependency set. The tracer receives
// SpanInfo with the operation name (e.g. "gerpo.GetFirst") and the bound table.
func ExampleWithTracer() {
	myTracer := func(ctx context.Context, span gerpo.SpanInfo) (context.Context, gerpo.SpanEnd) {
		// Open a span using your tracer of choice. Here we just stub the call.
		fmt.Println("start", span.Op, "table=", span.Table)
		return ctx, func(err error) {
			fmt.Println("end", span.Op, err)
		}
	}

	pool, _ := pgxpool.New(context.Background(), "postgres://localhost/db")

	repo, err := gerpo.New[User]().
		DB(pgx5.NewPoolAdapter(pool)).
		Table("users").
		Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
			c.Field(&m.ID).OmitOnUpdate()
			c.Field(&m.Name)
		}).
		WithTracer(myTracer).
		Build()
	if err != nil {
		panic(err)
	}
	_ = repo
}

// ExampleWithCacheStorage attaches the bundled context-scoped cache; reads
// inside a single context.Context get deduplicated, and any Insert/Update/
// Delete on the same repo invalidates the cache for that context.
func ExampleWithCacheStorage() {
	pool, _ := pgxpool.New(context.Background(), "postgres://localhost/db")
	cache := cachectx.New()

	repo, err := gerpo.New[User]().
		DB(pgx5.NewPoolAdapter(pool), executor.WithCacheStorage(cache)).
		Table("users").
		Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
			c.Field(&m.ID).OmitOnUpdate()
			c.Field(&m.Name)
		}).
		Build()
	if err != nil {
		panic(err)
	}

	// Wrap the request context once at the entry point so subsequent reads
	// served by this repo go through the cache.
	ctx := cachectx.WrapContext(context.Background())
	_, _ = repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
		h.Where().Field(&m.ID).EQ(uuid.UUID{})
	})
}
