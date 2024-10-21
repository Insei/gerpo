package api

import (
	"github.com/insei/gerpo/types"
)

type Core interface {
	ValidateFilters(string) error
	ValidateSorts(string) error
	GetAvailableFilters() map[string][]string
	GetAvailableSorts() []string
	ApplySorts(sorts string, target types.OrderTarget)
	ApplyFilters(filters string, target types.WhereTarget)
}
