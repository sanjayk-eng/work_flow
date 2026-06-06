package auth

import (
	"github/sanjay-khandelwal/internal/shared/core/database/postgres"
	"github/sanjay-khandelwal/internal/shared/security/password"

	"github.com/gin-gonic/gin"
)

// Module wires the auth layer together and registers routes.
// Call this once from main or a top-level router setup.
//
//	auth.New(db, hasher).Mount(router.Group("/api/v1"))
type Module struct {
	handler *Handler
}

func New(db *postgres.DB, hasher *password.Hasher) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, hasher)
	handler := NewHandler(svc)
	return &Module{handler: handler}
}

func (m *Module) Mount(rg *gin.RouterGroup) {
	RegisterRoutes(rg, m.handler)
}
