package query

import "fmt"

var (
	ErrApplyWhereClause         = fmt.Errorf("failed to apply WHERE clause")
	ErrApplyOrderByOperator     = fmt.Errorf("failed to apply ORDER BY operator")
	ErrApplyLimitOffsetOperator = fmt.Errorf("failed to apply LIMIT, OFFSET operator")
	ErrApplyJoinClause          = fmt.Errorf("failed to apply JOIN clause")
	ErrApplyGroupByClause       = fmt.Errorf("failed to apply GROUP BY operator")
	ErrApplyExcludeColumnRules  = fmt.Errorf("failed to apply exclude column rules")
)
