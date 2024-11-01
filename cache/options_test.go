package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithSource(t *testing.T) {
	testCases := []struct {
		name                 string
		sources              []Source
		expectedSourcesCount int
	}{
		{
			name:                 "With no source options",
			sources:              []Source{},
			expectedSourcesCount: 0,
		},
		{
			name:                 "With 1 source options",
			sources:              []Source{mockSource{}},
			expectedSourcesCount: 1,
		},
		{
			name:                 "With Multiple source options",
			sources:              []Source{mockSource{}},
			expectedSourcesCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := &modelBundle{}
			for _, source := range tc.sources {
				WithSource(source).apply(b)
			}
			assert.Equal(t, len(b.sources), tc.expectedSourcesCount)
		})
	}
}
