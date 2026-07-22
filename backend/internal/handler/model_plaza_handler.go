package handler

import (
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

func (h *ModelPlazaHandler) List(c *gin.Context) {
	entries, err := h.service.ListPublic(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, entries)
}
