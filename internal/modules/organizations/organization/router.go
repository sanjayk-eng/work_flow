package organization

import "github.com/gin-gonic/gin"

// RegisterRoutes wires all organization routes
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {

	org := r.Group("/organizations")
	// CRUD routes
	org.POST("", h.Create)
	org.GET("", h.List)
	org.GET("/:id", h.GetByID)
	org.GET("/slug/:slug", h.GetBySlug)
	org.PUT("/:id", h.Update)
	org.DELETE("/:id", h.Delete)
}
