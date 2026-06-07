package auth

import (
	"github/sanjay-khandelwal/internal/middleware"
	"github/sanjay-khandelwal/internal/shared/security/jwt"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtSvc *jwt.Service) {
	auth := rg.Group("/auth")

	// Public — no authentication required
	{
		auth.POST("/register", h.Register)
		auth.GET("/verify-email", h.EmailVerification)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
	}

	// Protected — valid access token required
	protected := auth.Group("", middleware.Authenticate(jwtSvc))
	{
		protected.POST("/logout", h.Logout)
		protected.POST("/logout-all", h.LogoutAll)
	}
}
