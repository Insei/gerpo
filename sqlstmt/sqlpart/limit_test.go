package sqlpart

import (
	"strconv"
	"testing"
)

func TestGetOffsetStr(t *testing.T) {
	tests := []struct {
		name   string
		offset uint64
		want   string
	}{
		{"zero offset", 0, ""},
		{"positive offset", 10, "10"},
		{"large offset", 123456789, "123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &LimitOffsetBuilder{}
			builder.SetOffset(tt.offset)

			got := builder.getOffsetStr()
			if got != tt.want {
				t.Errorf("getOffsetStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLimitOffsetBuilder_getLimitStr(t *testing.T) {
	tests := []struct {
		name   string
		limit  uint64
		expect string
	}{
		{
			name:   "Limit is zero",
			limit:  0,
			expect: "",
		},
		{
			name:   "Limit is non-zero",
			limit:  10,
			expect: "10",
		},
		{
			name:   "Limit is maximum uint64",
			limit:  ^uint64(0),
			expect: strconv.FormatUint(^uint64(0), 10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewLimitOffsetBuilder()
			builder.SetLimit(tt.limit)

			result := builder.getLimitStr()
			if result != tt.expect {
				t.Errorf("getLimitStr() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestLimitOffsetBuilder_SQL(t *testing.T) {
	tests := []struct {
		name   string
		limit  uint64
		offset uint64
		expect string
	}{
		{
			name:   "Both limit and offset are zero",
			limit:  0,
			offset: 0,
			expect: "",
		},
		{
			name:   "Only limit is set",
			limit:  10,
			offset: 0,
			expect: " LIMIT 10",
		},
		{
			name:   "Only offset is set",
			limit:  0,
			offset: 5,
			expect: " OFFSET 5",
		},
		{
			name:   "Both limit and offset are set",
			limit:  15,
			offset: 5,
			expect: " LIMIT 15 OFFSET 5",
		},
		{
			name:   "Maximum uint64 values",
			limit:  ^uint64(0),
			offset: ^uint64(0),
			expect: " LIMIT " + strconv.FormatUint(^uint64(0), 10) + " OFFSET " + strconv.FormatUint(^uint64(0), 10),
		},
		{
			name:   "Limit with leading/trailing spaces",
			limit:  20,
			offset: 10,
			expect: " LIMIT 20 OFFSET 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewLimitOffsetBuilder()
			builder.SetLimit(tt.limit)
			builder.SetOffset(tt.offset)

			result := builder.SQL()
			if result != tt.expect {
				t.Errorf("SQL() = %v, want %v", result, tt.expect)
			}
		})
	}
}
