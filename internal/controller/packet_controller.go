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
	packets, err := controller.Service.GetPackets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, packets)
}

func NewPacketController(service internal.PacketService) PacketController {
	return &PacketControllerImpl{Service: service}
}
