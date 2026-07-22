package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type ModelPlazaHandler struct {
	service *service.ModelPlazaService
}

func NewModelPlazaHandler(service *service.ModelPlazaService) *ModelPlazaHandler {
	return &ModelPlazaHandler{service: service}
}

type modelPlazaEntryRequest struct {
	ModelName   string   `json:"model_name" binding:"required,max=255"`
	DisplayName string   `json:"display_name" binding:"max=255"`
	Description string   `json:"description" binding:"max=2000"`
	GroupID     int64    `json:"group_id" binding:"required,gt=0"`
	Tags        []string `json:"tags"`
	SortOrder   int      `json:"sort_order"`
	Enabled     bool     `json:"enabled"`
}

func (r modelPlazaEntryRequest) toService() service.ModelPlazaEntryInput {
	return service.ModelPlazaEntryInput{
		ModelName: r.ModelName, DisplayName: r.DisplayName, Description: r.Description,
		GroupID: r.GroupID, Tags: r.Tags, SortOrder: r.SortOrder, Enabled: r.Enabled,
	}
}

func (h *ModelPlazaHandler) List(c *gin.Context) {
	entries, err := h.service.ListAdmin(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, entries)
}

func (h *ModelPlazaHandler) Create(c *gin.Context) {
	var req modelPlazaEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	entry, err := h.service.Create(c.Request.Context(), req.toService())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, entry)
}

func (h *ModelPlazaHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid model plaza entry ID")
		return
	}
	var req modelPlazaEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	entry, err := h.service.Update(c.Request.Context(), id, req.toService())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, entry)
}

func (h *ModelPlazaHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid model plaza entry ID")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Model removed from plaza"})
}
