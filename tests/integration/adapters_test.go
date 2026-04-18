//go:build integration

package integration

import (
	"testing"

	"github.com/insei/gerpo/executor"
	"github.com/insei/gerpo/executor/adapters/databasesql"
	"github.com/insei/gerpo/executor/adapters/pgx4"
	"github.com/insei/gerpo/executor/adapters/pgx5"
	"github.com/insei/gerpo/executor/adapters/placeholder"
)

// adapterBundle содержит один из поддерживаемых gerpo DBAdapter и его имя.
// Тесты получают bundle через forEachAdapter и не знают, какая реализация внутри.
type adapterBundle struct {
	name    string
	adapter executor.DBAdapter
}

// allAdapters возвращает конструкторы для всех трёх поддерживаемых адаптеров.
// Коннекты уже открыты в TestMain — здесь только оборачиваем их.
func allAdapters() []func() adapterBundle {
	return []func() adapterBundle{
		func() adapterBundle {
			return adapterBundle{name: "pgx5", adapter: pgx5.NewPoolAdapter(pgx5Pool)}
		},
		func() adapterBundle {
			return adapterBundle{name: "pgx4", adapter: pgx4.NewPoolAdapter(pgx4Pool)}
		},
		func() adapterBundle {
			return adapterBundle{
				name:    "databasesql",
				adapter: databasesql.NewAdapter(stdlibDB, databasesql.WithPlaceholder(placeholder.Dollar)),
			}
		},
	}
}

// forEachAdapter запускает fn как sub-test для каждого адаптера gerpo.
// Перед каждым подтестом таблицы очищаются, поэтому состояние между адаптерами
// не пересекается.
func forEachAdapter(t *testing.T, fn func(t *testing.T, ab adapterBundle)) {
	t.Helper()
	for _, make := range allAdapters() {
		ab := make()
		t.Run(ab.name, func(t *testing.T) {
			truncateAll(t)
			fn(t, ab)
		})
	}
}
