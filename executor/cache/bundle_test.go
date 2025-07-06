package cache

import (
	"context"
	"errors"
	"testing"
)

var _ Storage = &mockSource{}

type mockSource struct {
	result any
	err    error
}

func (m mockSource) Get(ctx context.Context, statement string, statementArgs ...any) (any, error) {
	return m.result, m.err
}

func (m mockSource) Clean(ctx context.Context) {}

func (m mockSource) Set(ctx context.Context, cache any, statement string, statementArgs ...any) {}

func TestModelBundle_Get(t *testing.T) {
	tests := []struct {
		name          string
		modelBundle   *storagesBundle
		expectedValue any
		expectedError error
	}{
		{
			name: "GetSuccess",
			modelBundle: &storagesBundle{
				storages: []Storage{
					&mockSource{
						result: "value1",
						err:    nil,
					},
				},
			},
			expectedValue: "value1",
			expectedError: nil,
		},
		{
			name: "GetError",
			modelBundle: &storagesBundle{
				storages: []Storage{
					&mockSource{
						err: errors.New("error occurred"),
					},
				},
			},
			expectedValue: nil,
			expectedError: ErrNotFound,
		},
		{
			name: "MultipleSources",
			modelBundle: &storagesBundle{
				storages: []Storage{
					&mockSource{
						err: errors.New("source 1 error"),
					},
					&mockSource{
						result: "value3",
						err:    nil,
					},
				},
			},
			expectedValue: "value3",
			expectedError: nil,
		},
	}

	ctx := context.Background()
	statement := "statement"
	var statementArgs []any

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.modelBundle.Get(ctx, statement, statementArgs...)
			if value != tt.expectedValue {
				t.Errorf("expected value %v but got %v", tt.expectedValue, value)
			}
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error %v but got %v", tt.expectedError, err)
			}
		})
	}
}

func TestModelBundle_Set(t *testing.T) {
	tests := []struct {
		name          string
		modelBundle   *storagesBundle
		setValue      any
		expectedError error
	}{
		{
			name: "SetSuccess",
			modelBundle: &storagesBundle{
				storages: []Storage{
					&mockSource{
						err: nil,
					},
				},
			},
			setValue:      "value1",
			expectedError: nil,
		},
		{
			name: "SetError",
			modelBundle: &storagesBundle{
				storages: []Storage{
					&mockSource{
						err: errors.New("set error"),
					},
				},
			},
			setValue:      "value2",
			expectedError: errors.New("set error"),
		},
		{
			name: "MultipleSources",
			modelBundle: &storagesBundle{
				storages: []Storage{
					&mockSource{
						err: errors.New("source 1 error"),
					},
					&mockSource{
						err: nil,
					},
				},
			},
			setValue:      "value3",
			expectedError: nil,
		},
	}

	ctx := context.Background()
	statement := "statement"
	var statementArgs []any

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.modelBundle.Set(ctx, tt.setValue, statement, statementArgs...)
		})
	}
}
func TestModelBundle_Clean(t *testing.T) {
	tests := []struct {
		name        string
		modelBundle *storagesBundle
	}{
		{
			name: "SingleSource",
			modelBundle: &storagesBundle{
				storages: []Storage{
					&mockSource{},
				},
			},
		},
		{
			name: "MultipleSources",
			modelBundle: &storagesBundle{
				storages: []Storage{
					&mockSource{},
					&mockSource{},
				},
			},
		},
		{
			name:        "NoSources",
			modelBundle: &storagesBundle{},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.modelBundle.Clean(ctx)
		})
	}
}

func TestNewModelBundle(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
	}{
		{
			name:    "WithoutOption",
			options: []Option{},
		},
		{
			name: "WithOneSource",
			options: []Option{
				WithStorage(&mockSource{}),
			},
		},
		{
			name: "WithMultipleSources",
			options: []Option{
				WithStorage(&mockSource{}),
				WithStorage(&mockSource{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewModelBundle(tt.options...).(*storagesBundle)
			if len(b.storages) != len(tt.options) {
				t.Errorf("expected number of storages %v but got %v", len(tt.options), len(b.storages))
			}
		})
	}
}
