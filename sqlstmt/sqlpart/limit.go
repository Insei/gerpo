package sqlpart

import (
	"fmt"
	"strconv"
	"strings"
)

type LimitOffset interface {
	GetOffset() uint64
	GetLimit() uint64
	SetOffset(offset uint64)
	SetLimit(size uint64)
}

type LimitOffsetBuilder struct {
	offset uint64
	limit  uint64
}

func NewLimitOffsetBuilder() *LimitOffsetBuilder {
	return &LimitOffsetBuilder{}
}

func (p *LimitOffsetBuilder) GetOffset() uint64 {
	return p.offset
}

func (p *LimitOffsetBuilder) GetLimit() uint64 {
	return p.limit
}

func (p *LimitOffsetBuilder) SetOffset(offset uint64) {
	p.offset = offset
}

func (p *LimitOffsetBuilder) SetLimit(size uint64) {
	p.limit = size
}

func (p *LimitOffsetBuilder) getOffsetStr() string {
	if p.offset == 0 {
		return ""
	}
	return strconv.FormatUint(p.GetOffset(), 10)
}

func (p *LimitOffsetBuilder) getLimitStr() string {
	if p.limit == 0 {
		return ""
	}
	return strconv.FormatUint(p.GetLimit(), 10)
}

func (p *LimitOffsetBuilder) SQL() string {
	sql := ""
	limitNumStr := p.getLimitStr()
	if strings.TrimSpace(limitNumStr) != "" {
		sql += fmt.Sprintf(" LIMIT %s", limitNumStr)
	}
	offsetNumStr := p.getOffsetStr()
	if strings.TrimSpace(offsetNumStr) != "" {
		sql += fmt.Sprintf(" OFFSET %s", offsetNumStr)
	}
	return sql
}
