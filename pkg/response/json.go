package response

import (
	"net/http"

	"github/sanjay-khandelwal/internal/shared/apperr"

	"github.com/gin-gonic/gin"
)

// Success writes a 200 JSON response with a data payload.
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// Created writes a 201 JSON response with a data payload.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

// NoContent writes a 204 response with no body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error maps an error to its HTTP status and writes a JSON error response.
// Uses apperr.HTTPStatus so typed AppErrors get the right code automatically.
func Error(c *gin.Context, err error) {
	status, msg := apperr.HTTPStatus(err)
	c.JSON(status, gin.H{"error": msg})
}

// Paginated writes a paginated response with data + pagination metadata.
func Paginated(c *gin.Context, data any, meta PaginationMeta) {
	c.JSON(http.StatusOK, gin.H{
		"data":       data,
		"pagination": meta,
	})
}

// PaginationMeta is embedded in every paginated response.
type PaginationMeta struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}
