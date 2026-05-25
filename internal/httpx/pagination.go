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

type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

func ParsePagination(c *gin.Context) Pagination {
	limit := parsePositiveInt(c.Query("limit"), DefaultLimit)
	if limit > MaxLimit {
		limit = MaxLimit
	}

	offset := parsePositiveInt(c.Query("offset"), -1)
	if offset == -1 {
		page := parsePositiveInt(c.Query("page"), 1)
		if page < 1 {
			page = 1
		}
		offset = (page - 1) * limit
	}

	if offset < 0 {
		offset = 0
	}

	return Pagination{Limit: limit, Offset: offset}
}

func (p Pagination) Page() int {
	if p.Limit < 1 {
		return 1
	}
	return p.Offset/p.Limit + 1
}

func BuildPaginationMeta(total int64, page, limit int) PaginationMeta {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return PaginationMeta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
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
