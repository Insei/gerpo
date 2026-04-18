//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	pgxv4 "github.com/jackc/pgx/v4/pgxpool"
	pgxv5 "github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const envDSN = "GERPO_INTEGRATION_DB_URL"

// Глобальные коннекты к БД — открываются один раз в TestMain и переиспользуются.
var (
	dsn       string
	pgx5Pool  *pgxv5.Pool
	pgx4Pool  *pgxv4.Pool
	stdlibDB  *sql.DB
)

func TestMain(m *testing.M) {
	dsn = os.Getenv(envDSN)
	if dsn == "" {
		fmt.Fprintf(os.Stderr, "skipping integration tests: %s is not set\n", envDSN)
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := openConnections(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to open connections: %v\n", err)
		os.Exit(1)
	}

	if err := applySchema(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to apply schema: %v\n", err)
		closeConnections()
		os.Exit(1)
	}

	code := m.Run()
	closeConnections()
	os.Exit(code)
}

func openConnections(ctx context.Context) error {
	var err error

	pgx5Pool, err = pgxv5.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("pgx5 pool: %w", err)
	}
	if err := pgx5Pool.Ping(ctx); err != nil {
		return fmt.Errorf("pgx5 ping: %w", err)
	}

	pgx4Pool, err = pgxv4.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("pgx4 pool: %w", err)
	}

	stdlibDB, err = sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("database/sql open: %w", err)
	}
	if err := stdlibDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database/sql ping: %w", err)
	}

	return nil
}

func closeConnections() {
	if pgx5Pool != nil {
		pgx5Pool.Close()
	}
	if pgx4Pool != nil {
		pgx4Pool.Close()
	}
	if stdlibDB != nil {
		_ = stdlibDB.Close()
	}
}

func applySchema(ctx context.Context) error {
	_, err := pgx5Pool.Exec(ctx, schemaSQL)
	return err
}

// truncateAll очищает все таблицы в зависимом порядке. Использовать перед каждым
// подтестом, чтобы сделать тесты независимыми друг от друга.
func truncateAll(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const stmt = `TRUNCATE TABLE comments, posts, users RESTART IDENTITY CASCADE`
	if _, err := pgx5Pool.Exec(ctx, stmt); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

// testCtx создаёт context.Background() с коротким таймаутом для одного теста.
func testCtx(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 10*time.Second)
}
