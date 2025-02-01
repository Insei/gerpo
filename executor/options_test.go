package executor

import (
	"testing"

	"github.com/insei/gerpo/cache"
)

func TestWithCacheBundle(t *testing.T) {
	tests := []struct {
		name   string
		source cache.Source
	}{
		{
			name:   "With Non-nil Cache Source",
			source: &MockCacheSource{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			option := WithCacheSource(test.source)
			opt := &options{}
			option.apply(opt)
			if opt.cacheSource != test.source {
				t.Errorf("expected %v, but got %v", test.source, opt.cacheSource)
			}
		})
	}
}
