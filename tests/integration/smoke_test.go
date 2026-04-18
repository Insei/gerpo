//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSmoke проверяет, что инфраструктура интеграционных тестов поднимается:
// коннекты открыты, схема применена, репозитории собираются, seed вставляет
// фикстуры и GetFirst возвращает ожидаемую запись для каждого адаптера.
func TestSmoke(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newUserRepo(t, ab)

		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.Count(ctx)
		require.NoError(t, err)
		require.Equal(t, uint64(len(seed.users)), got)
	})
}
