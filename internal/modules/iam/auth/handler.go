package auth

import (
	"github/sanjay-khandelwal/internal/modules/iam/auth/dto"
	"github/sanjay-khandelwal/internal/shared/apperr"
	"github/sanjay-khandelwal/pkg/response"
	"github/sanjay-khandelwal/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

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

	response.Success(c, gin.H{"message": "email verified successfully"})
}

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if ok, errs := validator.BindAndValidate(c, &req); !ok {
		response.ValidationError(c, errs)
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest("invalid request", err))
		return
	}

	res, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, res)
}

// Logout godoc
// POST /auth/logout
// Requires: Authorization: Bearer <access_token>
// Body:     { "refresh_token": "..." }
// Revokes the session tied to the provided refresh token.
func (h *Handler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest("refresh_token is required"))
		return
	}

	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "logged out successfully"})
}

// LogoutAll godoc
// POST /auth/logout-all
// Requires: Authorization: Bearer <access_token>
// Revokes every active session for the authenticated user.
func (h *Handler) LogoutAll(c *gin.Context) {
	rawID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.Unauthorized("unauthenticated"))
		return
	}

	userID, err := uuid.Parse(rawID.(string))
	if err != nil {
		response.Error(c, apperr.BadRequest("invalid user id"))
		return
	}

	if err := h.svc.LogoutAll(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "all sessions revoked"})
}
