//go:build benchpg

package bench

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const envDSN = "GERPO_BENCH_PG_DSN"

var (
	dsn  string
	pool *pgxpool.Pool
)

//go:embed schema.sql
var schemaSQL string

func TestMain(m *testing.M) {
	dsn = os.Getenv(envDSN)
	if dsn == "" {
		fmt.Fprintf(os.Stderr, "skipping bench: %s is not set\n", envDSN)
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var err error
	pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open pool: %v\n", err)
		os.Exit(1)
	}
	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ping: %v\n", err)
		pool.Close()
		os.Exit(1)
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		fmt.Fprintf(os.Stderr, "apply schema: %v\n", err)
		pool.Close()
		os.Exit(1)
	}

	code := m.Run()
	pool.Close()
	os.Exit(code)
}
