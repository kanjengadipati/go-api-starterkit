package httpx

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func ParsePagination(c *gin.Context) Pagination {
	limit := parsePositiveInt(c.Query("limit"), DefaultLimit)
	if limit > MaxLimit {
		limit = MaxLimit
	}

	offset := parsePositiveInt(c.Query("offset"), 0)
	return Pagination{Limit: limit, Offset: offset}
}

func (p Pagination) Page() int {
	if p.Limit < 1 {
		return 1
	}
	return p.Offset/p.Limit + 1
}

func parsePositiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}
