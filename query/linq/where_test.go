package linq

import (
	"testing"

	"github.com/insei/gerpo/types"
)

func TestWithCaseInsensitive(t *testing.T) {
	tests := []struct {
		name string
		op   types.Operation
		want types.Operation
	}{
		{
			name: "Convert",
			op:   types.OperationCT,
			want: types.OperationCT_IC,
		},
		{
			name: "Stay the same",
			op:   types.OperationGT,
			want: types.OperationGT,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := withIgnoreCase(tt.op); got != tt.want {
				t.Errorf("WithIgnoreCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
