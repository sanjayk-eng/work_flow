package middleware

import (
	"net/http"
	"strings"

	"github/sanjay-khandelwal/internal/shared/security/jwt"

	"github.com/gin-gonic/gin"
)

// Context keys set by middleware — use these constants in handlers
const (
	CtxUserID = "user_id"
	CtxEmail  = "email"
	CtxOrgID  = "org_id"
	CtxRole   = "role"
)

// Authenticate validates the access token and sets user_id in context.
// Pure authentication — no role, no org, no email.
//
// Flow: Bearer token → VerifyToken → set user_id → Next
func Authenticate(svc *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractBearer(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		claims, err := svc.VerifyToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxEmail, claims.Email)
		c.Next()
	}
}

func extractBearer(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
