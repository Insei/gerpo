package ctx

import (
	"testing"

	"github.com/insei/gerpo/logger"
	"github.com/stretchr/testify/assert"
)

type MockLogger struct {
	logger.Logger
}

// TestWithLogger tests WithLogger function
func TestWithLogger(t *testing.T) {
	mLogger := &MockLogger{}
	tableTests := []struct {
		name         string
		log          logger.Logger
		expectLogger logger.Logger
	}{
		{"With Nil Logger", nil, logger.NoopLogger},
		{"With Mock Logger", mLogger, mLogger},
	}

	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			source := &CtxCache{
				log: logger.NoopLogger,
			}
			opt := WithLogger(tt.log)
			opt.apply(source)

			assert.Equal(t, source.log, tt.expectLogger)
		})
	}
}

// TestWithKey tests WithKey function
func TestWithKey(t *testing.T) {
	tableTests := []struct {
		name      string
		key       string
		expectKey string
	}{
		{"With Non-empty Key", "testKey", "testKey"},
		{"With Empty Key", "", "defaultKey"},
	}

	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			source := &CtxCache{
				key: "defaultKey",
			}
			opt := WithKey(tt.key)
			opt.apply(source)

			assert.Equal(t, source.key, tt.expectKey)
		})
	}
}
