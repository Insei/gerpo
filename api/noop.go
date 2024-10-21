package api

import (
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

var (
	NoopAPIConnectorFactory Core    = &noopCore{}
	npAPIConnector          Applier = &noopAPIConnector{}
)

type noopCore struct {
}

func (n noopCore) GetAvailableFilters() map[string][]string {
	return map[string][]string{}
}

func (n noopCore) GetAvailableSorts() []string {
	return make([]string, 0)
}

type noopAPIConnector struct {
}

func (n *noopAPIConnector) ApplyFilters(_ string, _ types.WhereTarget) {
}

func (n *noopAPIConnector) ApplySorts(_ string, _ types.OrderTarget) {
}

func (n *noopAPIConnector) AppendFilters(_ string) {
}

func (n *noopAPIConnector) ApplyWhere(_ *sql.StringWhereBuilder) {
}

func (n noopCore) ValidateFilters(_ string) error {
	return nil
}

func (n noopCore) ValidateSorts(_ string) error {
	return nil
}

func (n noopCore) NewApplier() Applier {
	return npAPIConnector
}
