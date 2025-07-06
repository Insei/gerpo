package executor

import (
	"testing"

	"github.com/insei/gerpo/executor/cache"
)

func TestWithCacheBundle(t *testing.T) {
	tests := []struct {
		name    string
		storage cache.Storage
	}{
		{
			name:    "With Non-nil Cache Storage",
			storage: &MockCacheSource{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			option := WithCacheStorage(test.storage)
			opt := &options{}
			option.apply(opt)
			if opt.cacheSource != test.storage {
				t.Errorf("expected %v, but got %v", test.storage, opt.cacheSource)
			}
		})
	}
}
