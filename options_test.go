package gerpo

import (
	"context"
	"testing"

	"github.com/insei/gerpo/query"
)

type exampleModel struct {
	ID int
}

func TestWithBeforeInsert(t *testing.T) {
	tests := []struct {
		name          string
		existingFn    func(ctx context.Context, m *exampleModel)
		newFn         func(ctx context.Context, m *exampleModel)
		expectWrapped bool
	}{
		{
			name:          "Nil existing, non-nil new",
			existingFn:    nil,
			newFn:         func(ctx context.Context, m *exampleModel) {},
			expectWrapped: false,
		},
		{
			name:          "Non-nil existing, non-nil new",
			existingFn:    func(ctx context.Context, m *exampleModel) {},
			newFn:         func(ctx context.Context, m *exampleModel) {},
			expectWrapped: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repository[exampleModel]{
				beforeInsert: tt.existingFn,
			}
			opt := WithBeforeInsert(tt.newFn)
			opt.apply(r)
			if r.beforeInsert == nil && tt.newFn != nil {
				t.Errorf("beforeInsert is nil, want not nil")
			}
		})
	}
}

func TestWithBeforeUpdate(t *testing.T) {
	tests := []struct {
		name          string
		existingFn    func(ctx context.Context, m *exampleModel)
		newFn         func(ctx context.Context, m *exampleModel)
		expectWrapped bool
	}{
		{
			name:          "Nil existing, non-nil new",
			existingFn:    nil,
			newFn:         func(ctx context.Context, m *exampleModel) {},
			expectWrapped: false,
		},
		{
			name:          "Non-nil existing, non-nil new",
			existingFn:    func(ctx context.Context, m *exampleModel) {},
			newFn:         func(ctx context.Context, m *exampleModel) {},
			expectWrapped: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repository[exampleModel]{
				beforeUpdate: tt.existingFn,
			}
			opt := WithBeforeUpdate(tt.newFn)
			opt.apply(r)
			if r.beforeUpdate == nil && tt.newFn != nil {
				t.Errorf("beforeUpdate is nil, want not nil")
			}
		})
	}
}

func TestWithAfterSelect(t *testing.T) {
	tests := []struct {
		name       string
		existingFn func(ctx context.Context, m []*exampleModel)
		newFn      func(ctx context.Context, m []*exampleModel)
	}{
		{
			name:       "Nil existing, non-nil new",
			existingFn: nil,
			newFn:      func(ctx context.Context, m []*exampleModel) {},
		},
		{
			name:       "Non-nil existing, non-nil new",
			existingFn: func(ctx context.Context, m []*exampleModel) {},
			newFn:      func(ctx context.Context, m []*exampleModel) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repository[exampleModel]{
				afterSelect: tt.existingFn,
			}
			opt := WithAfterSelect(tt.newFn)
			opt.apply(r)
			if r.afterSelect == nil && tt.newFn != nil {
				t.Errorf("afterSelect is nil, want not nil")
			}
		})
	}
}

func TestWithAfterInsert(t *testing.T) {
	tests := []struct {
		name       string
		existingFn func(ctx context.Context, m *exampleModel)
		newFn      func(ctx context.Context, m *exampleModel)
	}{
		{
			name:       "Nil existing, non-nil new",
			existingFn: nil,
			newFn:      func(ctx context.Context, m *exampleModel) {},
		},
		{
			name:       "Non-nil existing, non-nil new",
			existingFn: func(ctx context.Context, m *exampleModel) {},
			newFn:      func(ctx context.Context, m *exampleModel) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repository[exampleModel]{
				afterInsert: tt.existingFn,
			}
			opt := WithAfterInsert(tt.newFn)
			opt.apply(r)
			if r.afterInsert == nil && tt.newFn != nil {
				t.Errorf("afterInsert is nil, want not nil")
			}
		})
	}
}

func TestWithAfterUpdate(t *testing.T) {
	tests := []struct {
		name       string
		existingFn func(ctx context.Context, m *exampleModel)
		newFn      func(ctx context.Context, m *exampleModel)
	}{
		{
			name:       "Nil existing, non-nil new",
			existingFn: nil,
			newFn:      func(ctx context.Context, m *exampleModel) {},
		},
		{
			name:       "Non-nil existing, non-nil new",
			existingFn: func(ctx context.Context, m *exampleModel) {},
			newFn:      func(ctx context.Context, m *exampleModel) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repository[exampleModel]{
				afterUpdate: tt.existingFn,
			}
			opt := WithAfterUpdate(tt.newFn)
			opt.apply(r)
			if r.afterUpdate == nil && tt.newFn != nil {
				t.Errorf("afterUpdate is nil, want not nil")
			}
		})
	}
}

func TestWithQuery(t *testing.T) {
	tests := []struct {
		name  string
		newFn func(m *exampleModel, h query.PersistentHelper[exampleModel])
		isNil bool
	}{
		{
			name:  "Non-nil newFn",
			newFn: func(m *exampleModel, h query.PersistentHelper[exampleModel]) {},
			isNil: false,
		},
		{
			name:  "Nil newFn",
			newFn: nil,
			isNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repository[exampleModel]{
				persistentQuery: query.NewPersistent[exampleModel](&exampleModel{}),
			}
			opt := WithQuery(tt.newFn)
			opt.apply(r)
			if !tt.isNil && r.persistentQuery == nil {
				t.Errorf("persistentQuery is nil, want not nil")
			}
		})
	}
}

func TestWithErrorTransformer(t *testing.T) {
	tests := []struct {
		name string
		fn   func(error) error
	}{
		{
			name: "Non-nil transformer",
			fn:   func(e error) error { return e },
		},
		{
			name: "Nil transformer",
			fn:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repository[exampleModel]{}
			opt := WithErrorTransformer[exampleModel](tt.fn)
			opt.apply(r)
			if tt.fn == nil && r.errorTransformer != nil {
				t.Errorf("errorTransformer should be nil")
			}
			if tt.fn != nil && r.errorTransformer == nil {
				t.Errorf("errorTransformer is nil, want not nil")
			}
		})
	}
}
