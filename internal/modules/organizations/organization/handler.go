package organization

import (
	"github/sanjay-khandelwal/internal/modules/organizations/organization/dto"
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

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateOrganizationRequest

	if ok, errs := validator.BindAndValidate(c, &req); !ok {
		response.ValidationError(c, errs)
		return
	}

	rawUserID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.Unauthorized("unauthenticated"))
		return
	}

	userID, err := uuid.Parse(rawUserID.(string))
	if err != nil {
		response.Error(c, apperr.BadRequest("invalid user id"))
		return
	}

	org, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, org)
}

func (h *Handler) GetByID(c *gin.Context) {

	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apperr.BadRequest("invalid organization id"))
		return
	}

	org, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, org)
}

func (h *Handler) GetBySlug(c *gin.Context) {

	slug := c.Param("slug")
	if slug == "" {
		response.Error(c, apperr.BadRequest("slug is required"))
		return
	}

	org, err := h.svc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, org)
}

func (h *Handler) List(c *gin.Context) {

	rawUserID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.Unauthorized("unauthenticated"))
		return
	}

	userID, err := uuid.Parse(rawUserID.(string))
	if err != nil {
		response.Error(c, apperr.BadRequest("invalid user id"))
		return
	}

	orgs, err := h.svc.ListByOwner(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, orgs)
}

func (h *Handler) Update(c *gin.Context) {

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apperr.BadRequest("invalid organization id"))
		return
	}

	var req dto.CreateOrganizationRequest

	if ok, errs := validator.BindAndValidate(c, &req); !ok {
		response.ValidationError(c, errs)
		return
	}

	org, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, org)
}

func (h *Handler) Delete(c *gin.Context) {

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apperr.BadRequest("invalid organization id"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "organization deleted successfully",
	})
}
