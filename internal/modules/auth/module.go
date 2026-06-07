package auth

import (
	"github/sanjay-khandelwal/internal/modules/user"
	"github/sanjay-khandelwal/internal/shared/core/config"
	"github/sanjay-khandelwal/internal/shared/core/database/postgres"
	"github/sanjay-khandelwal/internal/shared/security/jwt"
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

func New(db *postgres.DB, userservice user.Service, JWT config.JWTConfig) *Module {

	// password Hasher
	hashPassword := password.New()

	// JWT SERVICE
	jwt := jwt.New([]byte(JWT.SecretKey))

	// authRepository
	repo := NewRepository(db)

	// authService
	svc := NewService(repo, db, hashPassword, jwt, userservice)

	// authHandler
	handler := NewHandler(svc)
	return &Module{handler: handler}
}

func (m *Module) Mount(rg *gin.RouterGroup) {
	RegisterRoutes(rg, m.handler)
}
