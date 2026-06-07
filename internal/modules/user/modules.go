package user

import (
	"github/sanjay-khandelwal/internal/shared/core/database/postgres"
)

// Module wires the auth layer together and registers routes.
// Call this once from main or a top-level router setup.
//
//	user.New(db, hasher).Mount(router.Group("/api/v1"))
type Module struct {
	//Handler Handler
	Service Service
	Repo    Repository
}

func New(db *postgres.DB) *Module {

	repo := NewRepository(db)
	svc := NewService(repo)
	// h := NewHandler(svc)

	return &Module{
		//Handler: h,
		Service: svc,
		Repo:    repo,
	}
}

// func (m *Module) Mount(rg *gin.RouterGroup) {
// 	RegisterRoutes(rg, m)
// }
