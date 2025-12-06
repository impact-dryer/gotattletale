package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	internal "github.com/impact-dryer/gotattletale/internal/service"
)

type PacketController interface {
	GetPackets(c *gin.Context)
}

type PacketControllerImpl struct {
	Service internal.PacketService
}

func (controller *PacketControllerImpl) GetPackets(c *gin.Context) {
	limit := c.GetInt("limit")
	if limit <= 0 {
		limit = 100
	}

	sort := c.GetString("sort")
	if sort == "" {
		sort = "created_at"
	}
	packets, err := controller.Service.GetPackets(limit, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, packets)
}

func NewPacketController(service internal.PacketService) PacketController {
	return &PacketControllerImpl{Service: service}
}
