package iam

import (
	iamAuth "github/sanjay-khandelwal/internal/modules/iam/auth"
	"github/sanjay-khandelwal/internal/modules/iam/session"
	"github/sanjay-khandelwal/internal/modules/iam/user"
	"github/sanjay-khandelwal/internal/shared/core/config"
	"github/sanjay-khandelwal/internal/shared/core/database/postgres"
	"github/sanjay-khandelwal/internal/shared/security/jwt"
	"github/sanjay-khandelwal/internal/shared/security/password"

	"github.com/gin-gonic/gin"
)

// Module is the top-level IAM (Identity & Access Management) module.
// It owns auth, user, and session sub-modules and is the single
// thing main.go needs to instantiate for the entire identity domain.
type Module struct {
	authModule *iamAuth.Module
}

func New(db *postgres.DB, jwtCfg config.JWTConfig) *Module {
	// shared infrastructure
	hasher := password.New()
	jwtSvc := jwt.New([]byte(jwtCfg.SecretKey))

	// sub-module wiring
	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo)

	sessionRepo := session.NewRepository(db)
	sessionSvc := session.NewService(sessionRepo)

	authRepo := iamAuth.NewRepository(db)
	authSvc := iamAuth.NewService(authRepo, db, hasher, jwtSvc, userSvc, sessionSvc)
	authHandler := iamAuth.NewHandler(authSvc)

	return &Module{
		authModule: iamAuth.NewModule(authHandler, jwtSvc),
	}
}

func (m *Module) Mount(rg *gin.RouterGroup) {
	m.authModule.Mount(rg)
}
