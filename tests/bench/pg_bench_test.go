//go:build benchpg

package bench

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insei/gerpo"
	"github.com/insei/gerpo/executor/adapters/pgx5"
	"github.com/insei/gerpo/query"
)

// Бенчмарки ниже — парный портрет overhead'а gerpo поверх реальной PostgreSQL
// через pgx v5 pool. В отличие от tests/mock_bench_test.go, здесь в замер
// попадают сетевое IO, парсинг запроса на стороне PG и выполнение. Результаты
// показывают, насколько cost'ы слоя gerpo (SQL codegen, field mapping через
// fmap, cache-check при nil-storage) видны на фоне настоящего round-trip.
//
// Гоняется через pool, поднятый в TestMain. Сравнение остальных адаптеров
// (pgx4, database/sql) не добавляет полезного сигнала — latency round-trip в
// любом случае доминирует. Так что фиксируем один адаптер и один репозиторий.

type benchRow struct {
	ID        uuid.UUID
	Name      string
	Email     *string
	Age       int
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

const (
	pgBenchSeedRows = 1000
	pgBenchListSize = 20
)

// pgBenchContext holds everything a benchmark needs: a gerpo repo, the
// raw pool for direct SQL, a seeded ID to target in read/update, and a
// pre-cleared window for inserts.
type pgBenchContext struct {
	repo    gerpo.Repository[benchRow]
	seedID  uuid.UUID
	seedIDs []uuid.UUID
}

func setupPgBench(tb testing.TB) *pgBenchContext {
	tb.Helper()
	// TestMain already applied schema.sql (users only). Each bench starts from
	// a known state so results are comparable across runs.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, `TRUNCATE TABLE users RESTART IDENTITY`); err != nil {
		tb.Fatalf("truncate: %v", err)
	}

	ids := make([]uuid.UUID, pgBenchSeedRows)
	batch := make([][]any, 0, pgBenchSeedRows)
	createdAt := time.Now().UTC()
	for i := range ids {
		ids[i] = uuid.New()
		email := fmt.Sprintf("user-%d@bench.local", i)
		batch = append(batch, []any{ids[i], fmt.Sprintf("user-%d", i), email, 20 + i%50, createdAt})
	}
	// pgx CopyFrom is the fastest way to seed a few thousand rows.
	_, err := pool.CopyFrom(ctx,
		[]string{"users"},
		[]string{"id", "name", "email", "age", "created_at"},
		copyFromRows(batch),
	)
	if err != nil {
		tb.Fatalf("seed CopyFrom: %v", err)
	}

	adapter := pgx5.NewPoolAdapter(pool)
	repo, err := gerpo.New[benchRow]().
		Adapter(adapter).
		Table("users").
		Columns(func(m *benchRow, c *gerpo.ColumnBuilder[benchRow]) {
			c.Field(&m.ID).OmitOnUpdate()
			c.Field(&m.Name)
			c.Field(&m.Email)
			c.Field(&m.Age)
			c.Field(&m.CreatedAt).OmitOnUpdate()
			c.Field(&m.UpdatedAt).OmitOnInsert()
			c.Field(&m.DeletedAt).OmitOnInsert()
		}).
		Build()
	if err != nil {
		tb.Fatalf("build repo: %v", err)
	}

	return &pgBenchContext{
		repo:    repo,
		seedID:  ids[0],
		seedIDs: ids,
	}
}

// copyFromRows adapts a [][]any into pgx.CopyFromSource for the seed path.
type copyFromRowsIter struct {
	rows [][]any
	idx  int
}

func copyFromRows(rows [][]any) *copyFromRowsIter  { return &copyFromRowsIter{rows: rows} }
func (c *copyFromRowsIter) Next() bool             { return c.idx < len(c.rows) }
func (c *copyFromRowsIter) Values() ([]any, error) { r := c.rows[c.idx]; c.idx++; return r, nil }
func (c *copyFromRowsIter) Err() error             { return nil }

// scanBenchRow mirrors the column order the gerpo repo emits so Direct and
// Gerpo variants compare apples-to-apples.
func scanBenchRow(rows interface {
	Scan(dest ...any) error
}, r *benchRow) error {
	return rows.Scan(&r.ID, &r.Name, &r.Email, &r.Age, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt)
}

// ---------- GetFirst ----------

func BenchmarkPg_GetFirst_Direct(b *testing.B) {
	c := setupPgBench(b)
	ctx := context.Background()
	const sql = `SELECT id, name, email, age, created_at, updated_at, deleted_at FROM users WHERE id = $1 LIMIT 1`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := pool.Query(ctx, sql, c.seedID)
		if err != nil {
			b.Fatal(err)
		}
		var r benchRow
		if rows.Next() {
			if err := scanBenchRow(rows, &r); err != nil {
				b.Fatal(err)
			}
		}
		rows.Close()
	}
}

func BenchmarkPg_GetFirst_Gerpo(b *testing.B) {
	c := setupPgBench(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.repo.GetFirst(ctx, func(m *benchRow, h query.GetFirstHelper[benchRow]) {
			h.Where().Field(&m.ID).EQ(c.seedID)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------- GetList ----------

func BenchmarkPg_GetList_Direct(b *testing.B) {
	c := setupPgBench(b)
	_ = c
	ctx := context.Background()
	sql := fmt.Sprintf(`SELECT id, name, email, age, created_at, updated_at, deleted_at FROM users ORDER BY created_at DESC LIMIT %d`, pgBenchListSize)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := pool.Query(ctx, sql)
		if err != nil {
			b.Fatal(err)
		}
		out := make([]*benchRow, 0, pgBenchListSize)
		for rows.Next() {
			r := new(benchRow)
			if err := scanBenchRow(rows, r); err != nil {
				b.Fatal(err)
			}
			out = append(out, r)
		}
		rows.Close()
	}
}

func BenchmarkPg_GetList_Gerpo(b *testing.B) {
	c := setupPgBench(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.repo.GetList(ctx, func(m *benchRow, h query.GetListHelper[benchRow]) {
			h.OrderBy().Field(&m.CreatedAt).DESC()
			h.Page(1).Size(pgBenchListSize)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------- Count ----------

func BenchmarkPg_Count_Direct(b *testing.B) {
	_ = setupPgBench(b)
	ctx := context.Background()
	const sql = `SELECT count(*) FROM users WHERE age >= $1`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var n uint64
		row := pool.QueryRow(ctx, sql, 30)
		if err := row.Scan(&n); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPg_Count_Gerpo(b *testing.B) {
	c := setupPgBench(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.repo.Count(ctx, func(m *benchRow, h query.CountHelper[benchRow]) {
			h.Where().Field(&m.Age).GTE(30)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------- Insert ----------

func BenchmarkPg_Insert_Direct(b *testing.B) {
	_ = setupPgBench(b)
	ctx := context.Background()
	const sql = `INSERT INTO users (id, name, email, age, created_at) VALUES ($1, $2, $3, $4, $5)`
	createdAt := time.Now().UTC()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := uuid.New()
		email := "ins@bench.local"
		if _, err := pool.Exec(ctx, sql, id, "ins", email, 30, createdAt); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPg_Insert_Gerpo(b *testing.B) {
	c := setupPgBench(b)
	ctx := context.Background()
	createdAt := time.Now().UTC()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		email := "ins@bench.local"
		r := &benchRow{ID: uuid.New(), Name: "ins", Email: &email, Age: 30, CreatedAt: createdAt}
		if err := c.repo.Insert(ctx, r); err != nil {
			b.Fatal(err)
		}
	}
}

// ---------- Update ----------

func BenchmarkPg_Update_Direct(b *testing.B) {
	c := setupPgBench(b)
	ctx := context.Background()
	const sql = `UPDATE users SET name = $1, email = $2, age = $3 WHERE id = $4`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		email := "upd@bench.local"
		if _, err := pool.Exec(ctx, sql, "upd", email, 31, c.seedID); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPg_Update_Gerpo(b *testing.B) {
	c := setupPgBench(b)
	ctx := context.Background()
	email := "upd@bench.local"
	r := &benchRow{ID: c.seedID, Name: "upd", Email: &email, Age: 31}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.repo.Update(ctx, r, func(m *benchRow, h query.UpdateHelper[benchRow]) {
			h.Where().Field(&m.ID).EQ(c.seedID)
			h.Only(&m.Name, &m.Email, &m.Age)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------- Delete ----------
//
// Delete is measured against a non-matching id to keep every iteration
// identical — no seeding overhead creeps into the timed window. PG still
// parses + plans + executes, so the round-trip is honest; the only thing
// that doesn't happen is the row write. gerpo's Delete returns
// ErrNotFound when rowsAffected == 0, which we accept here.

func BenchmarkPg_Delete_Direct(b *testing.B) {
	_ = setupPgBench(b)
	ctx := context.Background()
	const sql = `DELETE FROM users WHERE id = $1`
	missing := uuid.New()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := pool.Exec(ctx, sql, missing); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPg_Delete_Gerpo(b *testing.B) {
	c := setupPgBench(b)
	ctx := context.Background()
	missing := uuid.New()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.repo.Delete(ctx, func(m *benchRow, h query.DeleteHelper[benchRow]) {
			h.Where().Field(&m.ID).EQ(missing)
		})
		if err != nil && !errors.Is(err, gerpo.ErrNotFound) {
			b.Fatal(err)
		}
	}
}

// TestComparePgxVsGerpo programmatically runs every Direct/Gerpo pair above
// and prints a summary table against real PostgreSQL. Companion to
// TestCompareDirectVsGerpo in tests/mock_bench_test.go — that one shows
// pure-layer overhead, this one shows what survives when real round-trip is
// in the mix.
//
// Skipped by default (takes ~30-60s against a local Postgres). Trigger:
//
//	GERPO_BENCH_REPORT=1 GERPO_INTEGRATION_DB_URL="postgres://..." \
//	    go test -tags=integration -run=TestComparePgxVsGerpo -v ./tests/integration/...
func TestComparePgxVsGerpo(t *testing.T) {
	if os.Getenv("GERPO_BENCH_REPORT") == "" {
		t.Skip("set GERPO_BENCH_REPORT=1 to run (~30-60s against real PG)")
	}

	suites := []struct {
		name   string
		direct func(b *testing.B)
		gerpo  func(b *testing.B)
	}{
		{"GetFirst", BenchmarkPg_GetFirst_Direct, BenchmarkPg_GetFirst_Gerpo},
		{"GetList", BenchmarkPg_GetList_Direct, BenchmarkPg_GetList_Gerpo},
		{"Count", BenchmarkPg_Count_Direct, BenchmarkPg_Count_Gerpo},
		{"Insert", BenchmarkPg_Insert_Direct, BenchmarkPg_Insert_Gerpo},
		{"Update", BenchmarkPg_Update_Direct, BenchmarkPg_Update_Gerpo},
		{"Delete", BenchmarkPg_Delete_Direct, BenchmarkPg_Delete_Gerpo},
	}

	type row struct {
		name                         string
		dNs, gNs                     int64
		dB, gB                       int64
		dAllocs, gAllocs             int64
		nsRatio, bRatio, allocsRatio float64
	}
	rows := make([]row, 0, len(suites))

	for _, s := range suites {
		d := testing.Benchmark(s.direct)
		g := testing.Benchmark(s.gerpo)
		rows = append(rows, row{
			name:        s.name,
			dNs:         d.NsPerOp(),
			gNs:         g.NsPerOp(),
			dB:          d.AllocedBytesPerOp(),
			gB:          g.AllocedBytesPerOp(),
			dAllocs:     d.AllocsPerOp(),
			gAllocs:     g.AllocsPerOp(),
			nsRatio:     safeDivPg(g.NsPerOp(), d.NsPerOp()),
			bRatio:      safeDivPg(g.AllocedBytesPerOp(), d.AllocedBytesPerOp()),
			allocsRatio: safeDivPg(g.AllocsPerOp(), d.AllocsPerOp()),
		})
	}

	var b strings.Builder
	b.WriteString("\npgx v5 (direct) vs gerpo on real PostgreSQL (lower is better; ratios show gerpo relative to direct)\n\n")
	fmt.Fprintf(&b, "%-9s  %12s  %12s  %7s  %10s  %10s  %7s  %8s  %8s  %7s\n",
		"Op", "Direct ns/op", "Gerpo ns/op", "× ns",
		"Direct B/op", "Gerpo B/op", "× B",
		"Direct allocs", "Gerpo allocs", "× allocs")
	b.WriteString(strings.Repeat("-", 120) + "\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "%-9s  %12d  %12d  %6.1fx  %10d  %10d  %6.1fx  %8d  %8d  %6.1fx\n",
			r.name, r.dNs, r.gNs, r.nsRatio,
			r.dB, r.gB, r.bRatio,
			r.dAllocs, r.gAllocs, r.allocsRatio)
	}
	b.WriteString(`
Real-PG view: network IO + PG planning dominate per-op latency, so the
ns ratios here reflect cost *as the caller experiences it*. Mock-adapter
ratios are in make bench-report — they isolate the framework layer.
`)
	t.Log(b.String())
}

func safeDivPg(a, b int64) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}
