package linq

import (
	"testing"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/stretchr/testify/assert"
)

type mockLimitOffset struct {
	sqlpart.LimitOffset
	limit  uint64
	offset uint64
}

func (m *mockLimitOffset) SetLimit(limit uint64) {
	m.limit = limit
}

func (m *mockLimitOffset) SetOffset(offset uint64) {
	m.offset = offset
}

type mockPaginationApplier struct {
	limitOffset sqlpart.LimitOffset
}

func (m *mockPaginationApplier) LimitOffset() sqlpart.LimitOffset {
	return m.limitOffset
}

func TestPaginationBuilder(t *testing.T) {
	testCases := []struct {
		name            string
		page            uint64
		size            uint64
		expectedLimit   uint64
		expectedOffset  uint64
		defaultLimit    uint64
		defaultOffest   uint64
		expectedPage    uint64
		expectedNewPage uint64
	}{
		{
			name:           "Default_pagination",
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "Set_page_1",
			page:           1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "Set_page_2",
			page:           2,
			expectedLimit:  10,
			expectedOffset: 10,
		},
		{
			name:           "Set_size_20",
			size:           20,
			expectedLimit:  20,
			expectedOffset: 0,
		},
		{
			name:           "Set_page_2_size_20",
			page:           2,
			size:           20,
			expectedLimit:  20,
			expectedOffset: 20,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewPaginationBuilder()

			if tc.page > 0 {
				builder.Page(tc.page)
			}
			if tc.size > 0 {
				builder.Size(tc.size)
			}

			mockLimitOffset := &mockLimitOffset{}
			applier := &mockPaginationApplier{
				limitOffset: mockLimitOffset,
			}
			builder.Apply(applier)

			assert.Equal(t, tc.expectedLimit, mockLimitOffset.limit)
			assert.Equal(t, tc.expectedOffset, mockLimitOffset.offset)
		})
	}
}
