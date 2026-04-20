package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insei/gerpo"
	extypes "github.com/insei/gerpo/executor/types"
	"github.com/insei/gerpo/query"
)

// Бенчмарки ниже сравнивают оверхед gerpo поверх DBAdapter (интерфейс адаптера
// драйвера) с прямыми вызовами того же адаптера на мок-БД. Реальная база
// здесь не участвует — цель замера в том, чтобы изолировать стоимость слоя
// gerpo (SQL codegen, маппинг, хуки, кеш) от сетевого IO.
//
// Direct-варианты моделируют ручной код, который сам пишет SQL и сам делает
// Scan. Gerpo-варианты проходят через Repository[User] с минимальной
// конфигурацией (без persistent/soft/virtual/hooks) — то же, что нужно для
// базовой CRUD-операции.

type benchUser struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt *time.Time
	Name      string
	DeletedAt *time.Time
}

// newBenchMockDB готовит mock-адаптер, возвращающий одну строку на QueryContext
// и результат с RowsAffected=1 на ExecContext — этого достаточно, чтобы gerpo
// и прямой код прошли happy path без специальной обработки.
func newBenchMockDB(rowsPerQuery int) *mockDB {
	db := newMockDB()
	db.QueryContextFn = func(ctx context.Context, q string, args ...any) (extypes.Rows, error) {
		return &mockRows{max: rowsPerQuery}, nil
	}
	db.ExecContextFn = func(ctx context.Context, q string, args ...any) (extypes.Result, error) {
		return &mockResult{rowsAffected: 1}, nil
	}
	return db
}

func newBenchRepo(b *testing.B, db *mockDB) gerpo.Repository[benchUser] {
	b.Helper()
	repo, err := gerpo.New[benchUser]().
		DB(db).
		Table("users").
		Columns(func(m *benchUser, c *gerpo.ColumnBuilder[benchUser]) {
			c.Field(&m.ID).OmitOnUpdate()
			c.Field(&m.CreatedAt).OmitOnUpdate()
			c.Field(&m.UpdatedAt).OmitOnInsert()
			c.Field(&m.Name)
			c.Field(&m.DeletedAt).OmitOnInsert()
		}).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	return repo
}

// scanUser — общий помощник для Direct-вариантов, эмулирует ручной Scan.
func scanUser(rows extypes.Rows, u *benchUser) error {
	return rows.Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt, &u.Name, &u.DeletedAt)
}

// ---------- GetFirst ----------

func BenchmarkGetFirst_Direct(b *testing.B) {
	db := newBenchMockDB(1)
	const sql = `SELECT users.id, users.created_at, users.updated_at, users.name, users.deleted_at FROM users WHERE users.id = $1 LIMIT 1`
	id := uuid.New()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := db.QueryContext(ctx, sql, id)
		if err != nil {
			b.Fatal(err)
		}
		var u benchUser
		if rows.Next() {
			if err := scanUser(rows, &u); err != nil {
				b.Fatal(err)
			}
		}
		_ = rows.Close()
	}
}

func BenchmarkGetFirst_Gerpo(b *testing.B) {
	db := newBenchMockDB(1)
	repo := newBenchRepo(b, db)
	id := uuid.New()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetFirst(ctx, func(m *benchUser, h query.GetFirstHelper[benchUser]) {
			h.Where().Field(&m.ID).EQ(id)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------- GetList ----------

const benchListSize = 10

func BenchmarkGetList_Direct(b *testing.B) {
	db := newBenchMockDB(benchListSize)
	const sql = `SELECT users.id, users.created_at, users.updated_at, users.name, users.deleted_at FROM users`
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := db.QueryContext(ctx, sql)
		if err != nil {
			b.Fatal(err)
		}
		out := make([]*benchUser, 0, benchListSize)
		for rows.Next() {
			u := new(benchUser)
			if err := scanUser(rows, u); err != nil {
				b.Fatal(err)
			}
			out = append(out, u)
		}
		_ = rows.Close()
	}
}

func BenchmarkGetList_Gerpo(b *testing.B) {
	db := newBenchMockDB(benchListSize)
	repo := newBenchRepo(b, db)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetList(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------- Count ----------

func BenchmarkCount_Direct(b *testing.B) {
	db := newBenchMockDB(1)
	const sql = `SELECT count(*) over() AS count FROM users WHERE users.name = $1 LIMIT 1`
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := db.QueryContext(ctx, sql, "alice")
		if err != nil {
			b.Fatal(err)
		}
		var c uint64
		if rows.Next() {
			if err := rows.Scan(&c); err != nil {
				b.Fatal(err)
			}
		}
		_ = rows.Close()
	}
}

func BenchmarkCount_Gerpo(b *testing.B) {
	db := newBenchMockDB(1)
	repo := newBenchRepo(b, db)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.Count(ctx, func(m *benchUser, h query.CountHelper[benchUser]) {
			h.Where().Field(&m.Name).EQ("alice")
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------- Insert ----------

func BenchmarkInsert_Direct(b *testing.B) {
	db := newBenchMockDB(0)
	const sql = `INSERT INTO users (id, created_at, name) VALUES ($1, $2, $3)`
	ctx := context.Background()
	u := benchUser{ID: uuid.New(), CreatedAt: time.Now().UTC(), Name: "alice"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.ExecContext(ctx, sql, u.ID, u.CreatedAt, u.Name)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInsert_Gerpo(b *testing.B) {
	db := newBenchMockDB(0)
	repo := newBenchRepo(b, db)
	ctx := context.Background()
	u := benchUser{ID: uuid.New(), CreatedAt: time.Now().UTC(), Name: "alice"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := repo.Insert(ctx, &u); err != nil {
			b.Fatal(err)
		}
	}
}

// ---------- Update ----------

func BenchmarkUpdate_Direct(b *testing.B) {
	db := newBenchMockDB(0)
	const sql = `UPDATE users SET name = $1 WHERE users.id = $2`
	ctx := context.Background()
	id := uuid.New()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := db.ExecContext(ctx, sql, "new-name", id)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := res.RowsAffected(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdate_Gerpo(b *testing.B) {
	db := newBenchMockDB(0)
	repo := newBenchRepo(b, db)
	ctx := context.Background()
	u := benchUser{ID: uuid.New(), CreatedAt: time.Now().UTC(), Name: "new-name"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.Update(ctx, &u, func(m *benchUser, h query.UpdateHelper[benchUser]) {
			h.Where().Field(&m.ID).EQ(u.ID)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------- Delete ----------

func BenchmarkDelete_Direct(b *testing.B) {
	db := newBenchMockDB(0)
	const sql = `DELETE FROM users WHERE users.id = $1`
	ctx := context.Background()
	id := uuid.New()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := db.ExecContext(ctx, sql, id)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := res.RowsAffected(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDelete_Gerpo(b *testing.B) {
	db := newBenchMockDB(0)
	repo := newBenchRepo(b, db)
	ctx := context.Background()
	id := uuid.New()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.Delete(ctx, func(m *benchUser, h query.DeleteHelper[benchUser]) {
			h.Where().Field(&m.ID).EQ(id)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestCompareDirectVsGerpo запускает программно пары Direct/Gerpo-бенчмарков
// и печатает сводную таблицу. Напоминание: все бенчи идут на mock-адаптере,
// поэтому IO = 0; то, что здесь выглядит как «×7 CPU», в реальной БД
// растворяется в сетевом IO запроса.
//
// По умолчанию скипается (бенчи занимают ~20 секунд). Явный запуск:
//
//	GERPO_BENCH_REPORT=1 go test -run=TestCompareDirectVsGerpo -v ./tests/
func TestCompareDirectVsGerpo(t *testing.T) {
	if os.Getenv("GERPO_BENCH_REPORT") == "" {
		t.Skip("set GERPO_BENCH_REPORT=1 to run (~20s)")
	}

	suites := []struct {
		name   string
		direct func(b *testing.B)
		gerpo  func(b *testing.B)
	}{
		{"GetFirst", BenchmarkGetFirst_Direct, BenchmarkGetFirst_Gerpo},
		{"GetList", BenchmarkGetList_Direct, BenchmarkGetList_Gerpo},
		{"Count", BenchmarkCount_Direct, BenchmarkCount_Gerpo},
		{"Insert", BenchmarkInsert_Direct, BenchmarkInsert_Gerpo},
		{"Update", BenchmarkUpdate_Direct, BenchmarkUpdate_Gerpo},
		{"Delete", BenchmarkDelete_Direct, BenchmarkDelete_Gerpo},
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
			nsRatio:     safeDiv(g.NsPerOp(), d.NsPerOp()),
			bRatio:      safeDiv(g.AllocedBytesPerOp(), d.AllocedBytesPerOp()),
			allocsRatio: safeDiv(g.AllocsPerOp(), d.AllocsPerOp()),
		})
	}

	var b strings.Builder
	b.WriteString("\nDirect vs Gerpo on mockDB (lower is better; ratios show gerpo relative to direct)\n\n")
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
Reminder: on a mock adapter IO = 0, so the "×" columns are a pure-overhead
portrait of the gerpo layer. Against a real PostgreSQL the per-op gerpo
cost (~0.5-1.5 µs) gets absorbed by network latency — README reports
only +8% ns on a real pgx v4 pool.
`)
	t.Log(b.String())
}

func safeDiv(a, b int64) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}
