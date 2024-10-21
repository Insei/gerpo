package api

import (
	"github.com/insei/gerpo/types"
)

type Applier interface {
	FiltersApplier
	ApplySorts(sorts string, target types.OrderTarget)
}

type FiltersApplier interface {
	ApplyFilters(filters string, target types.WhereTarget)
}

type Core interface {
	ValidateFilters(string) error
	ValidateSorts(string) error
	GetAvailableFilters() map[string][]string
	GetAvailableSorts() []string
	NewApplier() Applier
}
