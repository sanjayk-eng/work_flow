package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Page holds parsed offset-based pagination params.
type Page struct {
	Page     int
	PageSize int
}

// Offset returns the SQL OFFSET value for the current page.
func (p Page) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit returns the SQL LIMIT value.
func (p Page) Limit() int {
	return p.PageSize
}

// ParsePage reads ?page= and ?page_size= from the request query string.
// Falls back to defaults if missing or invalid. Clamps page_size to MaxPageSize.
func ParsePage(c *gin.Context) Page {
	page := parseIntQuery(c, "page", DefaultPage)
	pageSize := parseIntQuery(c, "page_size", DefaultPageSize)

	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return Page{Page: page, PageSize: pageSize}
}

// Cursor holds cursor-based pagination state for keyset pagination.
type Cursor struct {
	After    string // opaque cursor value (e.g. base64-encoded ID or timestamp)
	PageSize int
}

// ParseCursor reads ?after= and ?page_size= from the request query string.
func ParseCursor(c *gin.Context) Cursor {
	after := c.Query("after")
	pageSize := parseIntQuery(c, "page_size", DefaultPageSize)

	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return Cursor{After: after, PageSize: pageSize}
}

// TotalPages calculates the number of pages given a total record count.
func TotalPages(total, pageSize int) int {
	if pageSize == 0 {
		return 0
	}
	pages := total / pageSize
	if total%pageSize != 0 {
		pages++
	}
	return pages
}

func parseIntQuery(c *gin.Context, key string, defaultVal int) int {
	raw := c.Query(key)
	if raw == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return defaultVal
	}
	return val
}
