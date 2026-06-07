package auth

import (
	"github/sanjay-khandelwal/internal/modules/auth/dto"
	"github/sanjay-khandelwal/internal/shared/apperr"
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

	if ok, errs := validator.BindAndValidate(c, &req); !ok {
		response.ValidationError(c, errs)
		return
	}

	user, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

func (h *Handler) EmailVerification(c *gin.Context) {

	token := c.Query("token")
	if token == "" {
		response.Error(c, apperr.BadRequest("token is required"))
		return
	}

	if err := h.svc.EmailVerify(c.Request.Context(), token); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "email verified successfully",
	})
}
