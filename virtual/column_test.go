package virtual

import (
	"context"
	"testing"

	"github.com/insei/fmap/v3"
	"github.com/stretchr/testify/assert"

	"github.com/insei/gerpo/types"
)

func TestColumnGetAvailableFilterOperations(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Active")

	c := column{
		base: &types.ColumnBase{
			Filters: types.NewFilterManagerForField(field),
		},
	}

	t.Run("Test GetAvailableFilterOperations", func(t *testing.T) {
		c.GetAvailableFilterOperations()
	})
}

func TestColumnIsAvailableFilterOperation(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Active")

	c := column{
		base: &types.ColumnBase{
			Filters: types.NewFilterManagerForField(field),
		},
	}

	t.Run("Test GetAvailableFilterOperations", func(t *testing.T) {
		c.IsAvailableFilterOperation(types.OperationIN)
	})
}

func TestColumnGetFilterFn(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("Active")

	c := column{
		base: &types.ColumnBase{
			Filters: types.NewFilterManagerForField(field),
		},
	}

	t.Run("Test GetFilterFn", func(t *testing.T) {
		c.GetFilterFn(types.OperationIN)
	})
}

func TestColumnIsAllowedAction(t *testing.T) {
	c := &column{
		base: &types.ColumnBase{
			AllowedActions: []types.SQLAction{types.SQLActionSelect, types.SQLActionInsert},
		},
	}

	testCases := []struct {
		action   types.SQLAction
		expected bool
		name     string
	}{
		{
			name:     "SQLActionSelect should be allowed",
			action:   types.SQLActionSelect,
			expected: true,
		},
		{
			name:     "SQLActionInsert should be allowed",
			action:   types.SQLActionInsert,
			expected: true,
		},
		{
			name:     "SQLActionUpdate should not be allowed",
			action:   types.SQLActionUpdate,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ok := c.IsAllowedAction(tc.action)
			assert.Equal(t, tc.expected, ok)
		})
	}
}

func TestColumnToSQL(t *testing.T) {
	c := column{
		base: &types.ColumnBase{
			ToSQL: func(ctx context.Context) string {
				return "test"
			},
		},
	}

	t.Run("Test ToSQL", func(t *testing.T) {
		sql := c.ToSQL(context.Background())
		assert.Equal(t, "test", sql)
	})
}

func TestColumnGetPtr(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("NonBool")

	c := column{
		base: &types.ColumnBase{
			GetPtr: func(model any) any {
				return field.GetPtr(model)
			},
		},
	}

	model := &TestModel{NonBool: ""}

	t.Run("Test GetPtr", func(t *testing.T) {
		ptr := c.GetPtr(model)
		assert.NotNil(t, ptr)
	})
}

func TestColumnGetField(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("NonBool")

	c := column{
		base: &types.ColumnBase{
			Field: field,
		},
	}

	t.Run("Test GetField", func(t *testing.T) {
		field2 := c.GetField()
		assert.Equal(t, field, field2)
	})
}

func TestColumnGetAllowedActions(t *testing.T) {
	allowedActions := []types.SQLAction{types.SQLActionSelect, types.SQLActionInsert}

	c := &column{
		base: &types.ColumnBase{
			AllowedActions: allowedActions,
		},
	}

	t.Run("Test GetAllowedActions", func(t *testing.T) {
		actions := c.GetAllowedActions()
		assert.Equal(t, allowedActions, actions)
	})
}

func TestColumnName(t *testing.T) {
	c := &column{}

	t.Run("Test name returns default values", func(t *testing.T) {
		name, ok := c.Name()
		assert.Equal(t, "", name)
		assert.False(t, ok)
	})
}

func TestColumnTable(t *testing.T) {
	c := &column{}

	t.Run("Test table returns default values", func(t *testing.T) {
		name, ok := c.Table()
		assert.Equal(t, "", name)
		assert.False(t, ok)
	})
}

func TestColumnNew(t *testing.T) {
	fields, _ := fmap.Get[TestModel]()
	field := fields.MustFind("NonBool")

	c, err := New(field)

	t.Run("Test New column initialization", func(t *testing.T) {
		assert.Nil(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, []types.SQLAction{types.SQLActionSelect}, c.GetAllowedActions())
	})
}
