package auth

import (
	"github/sanjay-khandelwal/internal/middleware"
	"github/sanjay-khandelwal/internal/shared/security/jwt"

	"github.com/gin-gonic/gin"
)

// Module holds the wired auth handler and exposes Mount.
// Wiring (repo → service → handler) is done by the parent iam.Module.
type Module struct {
	handler *Handler
	jwtSvc  *jwt.Service
}

func NewModule(handler *Handler, jwtSvc *jwt.Service) *Module {
	return &Module{handler: handler, jwtSvc: jwtSvc}
}

func (m *Module) Mount(rg *gin.RouterGroup) {
	RegisterRoutes(rg, m.handler, m.jwtSvc)
}

// authMiddleware returns the Authenticate middleware bound to this module's jwt service.
func (m *Module) authMiddleware() gin.HandlerFunc {
	return middleware.Authenticate(m.jwtSvc)
}
