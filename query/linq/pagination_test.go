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
		expectErr       bool
	}{
		{
			name:           "Default_pagination",
			expectedLimit:  0,
			expectedOffset: 0,
		},
		{
			name:           "Set_page_1_size_0",
			page:           1,
			size:           0,
			expectedLimit:  0,
			expectedOffset: 0,
			expectErr:      true,
		},
		{
			name:           "Set_page_2_size_0",
			page:           2,
			size:           0,
			expectedLimit:  0,
			expectedOffset: 0,
			expectErr:      true,
		},
		{
			name:           "Set_page_1_size_10",
			page:           1,
			size:           10,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "Set_page_2_size_10",
			page:           2,
			size:           10,
			expectedLimit:  10,
			expectedOffset: 10,
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
			err := builder.Apply(applier)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedLimit, mockLimitOffset.limit)
			assert.Equal(t, tc.expectedOffset, mockLimitOffset.offset)
		})
	}
}
