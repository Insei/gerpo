package executor

import (
	"testing"

	"github.com/insei/gerpo/cache"
)

func TestWithCacheBundle(t *testing.T) {
	tests := []struct {
		name        string
		cacheBundle cache.Source
	}{
		{
			name:        "With Non-nil CacheBundle",
			cacheBundle: &MockModelBundle{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			option := WithCacheBundle(test.cacheBundle)
			opt := &options{}
			option.apply(opt)
			if opt.cacheBundle != test.cacheBundle {
				t.Errorf("expected %v, but got %v", test.cacheBundle, opt.cacheBundle)
			}
		})
	}
}
