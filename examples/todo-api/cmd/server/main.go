// Binary server is a minimal production-shaped gerpo example: PostgreSQL
// pool, goose migrations run on boot, one CRUD repository for tasks, request-
// scope cache wired as HTTP middleware, graceful shutdown.
package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/insei/gerpo/executor/adapters/pgx5"
	cachectx "github.com/insei/gerpo/executor/cache/ctx"

	"github.com/insei/gerpo/examples/todo-api/internal/task"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func main() {
	if err := run(); err != nil {
		log.Fatalf("todo-api: %v", err)
	}
}

func run() error {
	dsn := env("DATABASE_URL", "postgres://todo:todo@localhost:5432/todo?sslmode=disable")
	addr := env("HTTP_ADDR", ":8080")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := runMigrations(ctx, dsn); err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	adapter := pgx5.NewPoolAdapter(pool)
	repo, err := task.NewRepository(adapter)
	if err != nil {
		return err
	}
	svc := task.NewService(repo, adapter)
	handler := task.NewHandler(svc)

	mux := http.NewServeMux()
	handler.Register(mux)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           cacheMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}

// cacheMiddleware wraps the incoming request context so every repository call
// downstream shares the request-scope dedup cache and observes auto-
// invalidation on writes.
func cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := cachectx.WrapContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// runMigrations applies every migration in the embedded FS before the pool
// opens for business. goose's internal bookkeeping table `goose_db_version`
// stays in the same database, so migrations are idempotent across restarts.
func runMigrations(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	goose.SetBaseFS(migrationFS)
	goose.SetLogger(goose.NopLogger())
	return goose.UpContext(ctx, db, "migrations")
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
