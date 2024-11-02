package executor

import (
	"testing"

	"github.com/insei/gerpo/cache"
	"github.com/insei/gerpo/logger"
)

func TestWithCacheBundle(t *testing.T) {
	tests := []struct {
		name        string
		cacheBundle cache.ModelBundle
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

type MockLogger struct {
	logger.Logger
}

func TestWithLogger(t *testing.T) {
	tests := []struct {
		name   string
		logger logger.Logger
	}{
		{
			name:   "With Non-nil Logger",
			logger: &MockLogger{},
		},
		{
			name:   "With nil Logger",
			logger: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			option := WithLogger(test.logger)
			opt := &options{}
			option.apply(opt)
			if opt.log != test.logger {
				t.Errorf("expected %v, but got %v", test.logger, opt.log)
			}
		})
	}
}
