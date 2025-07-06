package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithSource(t *testing.T) {
	testCases := []struct {
		name                 string
		sources              []Storage
		expectedSourcesCount int
	}{
		{
			name:                 "With no source options",
			sources:              []Storage{},
			expectedSourcesCount: 0,
		},
		{
			name:                 "With 1 source options",
			sources:              []Storage{mockSource{}},
			expectedSourcesCount: 1,
		},
		{
			name:                 "With Multiple source options",
			sources:              []Storage{mockSource{}},
			expectedSourcesCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := &storagesBundle{}
			for _, source := range tc.sources {
				WithStorage(source).apply(b)
			}
			assert.Equal(t, len(b.storages), tc.expectedSourcesCount)
		})
	}
}
