package auth

import (
	"net/http"

	"github/sanjay-khandelwal/internal/modules/auth/dto"
	"github/sanjay-khandelwal/internal/shared/core/logger"
	"github/sanjay-khandelwal/pkg/response"
	"github/sanjay-khandelwal/pkg/validator"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Register godoc
// POST /auth/register
func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	// bind JSON + validate in one call
	if ok, errs := validator.BindAndValidate(c, &req); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errs})
		return
	}

	log := logger.FromContext(c.Request.Context())
	log.Info("auth: register attempt", "email", req.Email)

	res, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	log.Info("auth: user registered", "user_id", res.User.ID)
	response.Created(c, res)
}
