package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireRole checks that the role resolved by OrgContext middleware
// matches one of the allowed roles.
// Must be used AFTER OrgContext middleware (requires role in context).
//
// Flow: role (from context) → match allowedRoles → allow or 403
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(CtxRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no role in context"})
			return
		}

		if !hasRole(role.(string), allowedRoles) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}

func hasRole(role string, allowed []string) bool {
	for _, r := range allowed {
		if strings.EqualFold(role, r) {
			return true
		}
	}
	return false
}
