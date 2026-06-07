package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// OrgMemberResolver is implemented by the organization repository.
// Middleware depends on this interface — not on the concrete DB layer.
type OrgMemberResolver interface {
	// GetMemberRole returns the role of userID inside orgID.
	// Returns ("", ErrNotFound) if the user is not a member.
	GetMemberRole(ctx context.Context, userID, orgID string) (string, error)
}

// OrgContext resolves the caller's role inside an organization dynamically via DB.
// Must be used AFTER Authenticate middleware (requires user_id in context).
//
// org_id is read from the request header "X-Org-ID".
//
// Flow: user_id (from JWT) + org_id (from header) → DB membership lookup → role set in context
func OrgContext(resolver OrgMemberResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, ok := c.Get(CtxUserID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}
		userID, ok := val.(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}

		orgID := c.GetHeader("X-Org-ID")
		if orgID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing X-Org-ID header"})
			return
		}

		role, err := resolver.GetMemberRole(c.Request.Context(), userID, orgID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
			return
		}

		c.Set(CtxOrgID, orgID)
		c.Set(CtxRole, role)
		c.Next()
	}
}
