package organization

import (
	"github/sanjay-khandelwal/internal/shared/core/database/postgres"
)

type Module struct {
	Handler *Handler
}

func New(db *postgres.DB) *Module {

	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	return &Module{
		Handler: h,
	}
}
