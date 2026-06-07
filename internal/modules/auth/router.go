package auth

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts all auth routes onto the provided router group.
//
// Public routes (no auth middleware):
//
//	POST /auth/register
func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.GET("/verify-email", h.EmailVerification)
		
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		// auth.POST("/logout", h.Logout) // NEW (important)

		// auth.GET("/me", middleware.Auth(), h.Me) // NEW (identity check)
	}
}
