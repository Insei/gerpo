//go:build integration

package integration

import (
	"testing"

	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWhere_EQ — точное совпадение по uuid и строке.
func TestWhere_EQ(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetList(ctx, func(m *Post, h query.GetListHelper[Post]) {
			h.Where().Field(&m.ID).EQ(seed.posts[3].ID)
		})
		require.NoError(t, err)
		assert.Len(t, got, 1)
	})
}

// TestWhere_NEQ — отрицание: все, кроме одной записи.
func TestWhere_NEQ(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.GetList(ctx, func(m *Post, h query.GetListHelper[Post]) {
			h.Where().Field(&m.ID).NEQ(seed.posts[0].ID)
		})
		require.NoError(t, err)
		assert.Len(t, got, len(seed.posts)-1)
	})
}

// TestWhere_LT_LTE_GT_GTE — числовые операторы на поле age пользователя.
// users 0..9 имеют age 20..29 соответственно.
func TestWhere_LT_LTE_GT_GTE(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		cases := []struct {
			name  string
			query func(m *User, h query.CountHelper[User])
			want  uint64
		}{
			{"LT 23", func(m *User, h query.CountHelper[User]) { h.Where().Field(&m.Age).LT(23) }, 3},   // 20,21,22
			{"LTE 23", func(m *User, h query.CountHelper[User]) { h.Where().Field(&m.Age).LTE(23) }, 4}, // 20..23
			{"GT 25", func(m *User, h query.CountHelper[User]) { h.Where().Field(&m.Age).GT(25) }, 4},   // 26..29
			{"GTE 25", func(m *User, h query.CountHelper[User]) { h.Where().Field(&m.Age).GTE(25) }, 5}, // 25..29
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := repo.Count(ctx, c.query)
				require.NoError(t, err)
				assert.Equal(t, c.want, got)
			})
		}
	})
}

// TestWhere_IN_NIN — проверка IN/NIN по набору uuid.
func TestWhere_IN_NIN(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		wanted := []any{seed.posts[0].ID, seed.posts[5].ID, seed.posts[10].ID}

		in, err := repo.GetList(ctx, func(m *Post, h query.GetListHelper[Post]) {
			h.Where().Field(&m.ID).IN(wanted...)
		})
		require.NoError(t, err)
		assert.Len(t, in, 3)

		nin, err := repo.GetList(ctx, func(m *Post, h query.GetListHelper[Post]) {
			h.Where().Field(&m.ID).NIN(wanted...)
		})
		require.NoError(t, err)
		assert.Len(t, nin, len(seed.posts)-3)
	})
}

// TestWhere_EQ_Nil / NEQ_Nil — генерация IS NULL / IS NOT NULL для nullable *string.
// users: у чётных индексов есть email, у нечётных — NULL.
func TestWhere_EQ_Nil(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.Count(ctx, func(m *User, h query.CountHelper[User]) {
			h.Where().Field(&m.Email).EQ(nil)
		})
		require.NoError(t, err)
		assert.Equal(t, uint64(5), got, "5 users have NULL email")
	})
}

func TestWhere_NEQ_Nil(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newUserRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		got, err := repo.Count(ctx, func(m *User, h query.CountHelper[User]) {
			h.Where().Field(&m.Email).NEQ(nil)
		})
		require.NoError(t, err)
		assert.Equal(t, uint64(5), got, "5 users have non-NULL email")
	})
}

// TestWhere_LIKE_Operators — CT/NCT/BW/NBW/EW/NEW против поля Title.
// Формат "Post N" для N = 0..29.
func TestWhere_LIKE_Operators(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		cases := []struct {
			name string
			op   func(m *Post, h query.CountHelper[Post])
			want uint64
		}{
			// CT("1") cot совпадает с: 1, 10..19, 21 → 12 записей.
			{"CT '1'", func(m *Post, h query.CountHelper[Post]) { h.Where().Field(&m.Title).CT("1") }, 12},
			// NCT("1") — все кроме тех 12 → 18.
			{"NCT '1'", func(m *Post, h query.CountHelper[Post]) { h.Where().Field(&m.Title).NCT("1") }, 18},
			// BW("Post 1") — префикс "Post 1": "Post 1", "Post 10..19" = 11.
			{"BW 'Post 1'", func(m *Post, h query.CountHelper[Post]) { h.Where().Field(&m.Title).BW("Post 1") }, 11},
			// NBW("Post 1") — 30 - 11 = 19.
			{"NBW 'Post 1'", func(m *Post, h query.CountHelper[Post]) { h.Where().Field(&m.Title).NBW("Post 1") }, 19},
			// EW("9") — "Post 9", "Post 19", "Post 29" = 3.
			{"EW '9'", func(m *Post, h query.CountHelper[Post]) { h.Where().Field(&m.Title).EW("9") }, 3},
			// NEW("9") — 30 - 3 = 27.
			{"NEW '9'", func(m *Post, h query.CountHelper[Post]) { h.Where().Field(&m.Title).NEW("9") }, 27},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := repo.Count(ctx, c.op)
				require.NoError(t, err)
				assert.Equal(t, c.want, got)
			})
		}
	})
}

// TestWhere_LIKE_IgnoreCase — регистронезависимые варианты LIKE-операторов.
func TestWhere_LIKE_IgnoreCase(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		caseSensitive, err := repo.Count(ctx, func(m *Post, h query.CountHelper[Post]) {
			h.Where().Field(&m.Title).CT("post")
		})
		require.NoError(t, err)
		assert.Equal(t, uint64(0), caseSensitive, "CT is case-sensitive by default")

		caseInsensitive, err := repo.Count(ctx, func(m *Post, h query.CountHelper[Post]) {
			h.Where().Field(&m.Title).CT("post", true)
		})
		require.NoError(t, err)
		assert.Equal(t, uint64(30), caseInsensitive, "CT(…, true) matches regardless of case")
	})
}

// TestWhere_AND — два последовательных Field-условия склеиваются через AND автоматически.
// Заметка: gerpo генерирует `LIKE CONCAT('%', ?, '%')` для CT/NCT/BW/EW, что ломается
// в PostgreSQL, когда в запросе есть параметры разных типов — PG не может вывести тип
// из CONCAT без явного cast. Поэтому в AND/OR/Group используем однотипные примитивы
// (bool, uuid, int), а LIKE-семейство тестируется отдельно в TestWhere_LIKE_Operators.
func TestWhere_AND(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		// Published=true AND UserID=users[0].ID. User[0] владеет постами 0,1,2.
		// Из них published у чётных индексов: 0 и 2 → 2 записи.
		got, err := repo.Count(ctx, func(m *Post, h query.CountHelper[Post]) {
			h.Where().Field(&m.Published).EQ(true).
				AND().Field(&m.UserID).EQ(seed.users[0].ID)
		})
		require.NoError(t, err)
		assert.Equal(t, uint64(2), got)
	})
}

// TestWhere_OR — OR между двумя полями.
func TestWhere_OR(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		// Published=true (15 постов) OR UserID=users[9].ID (3 поста: 27,28,29;
		// из них published — только 28). Уникальных: 15 published + 27 + 29 = 17.
		got, err := repo.Count(ctx, func(m *Post, h query.CountHelper[Post]) {
			h.Where().Field(&m.Published).EQ(true).
				OR().Field(&m.UserID).EQ(seed.users[9].ID)
		})
		require.NoError(t, err)
		assert.Equal(t, uint64(17), got)
	})
}

// TestWhere_Group — группировка скобками меняет приоритет логики.
func TestWhere_Group(t *testing.T) {
	forEachAdapter(t, func(t *testing.T, ab adapterBundle) {
		seed := defaultSeed(t)
		repo := newPostRepo(t, ab)
		ctx, cancel := testCtx(t)
		defer cancel()

		// (UserID=users[0].ID OR UserID=users[1].ID) AND Published=true.
		// User[0] posts=0,1,2 (published: 0,2). User[1] posts=3,4,5 (published: 4). Итого 3.
		got, err := repo.Count(ctx, func(m *Post, h query.CountHelper[Post]) {
			h.Where().Group(func(t types.WhereTarget) {
				t.Field(&m.UserID).EQ(seed.users[0].ID).
					OR().Field(&m.UserID).EQ(seed.users[1].ID)
			}).AND().Field(&m.Published).EQ(true)
		})
		require.NoError(t, err)
		assert.Equal(t, uint64(3), got)
	})
}
