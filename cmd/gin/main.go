package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/impact-dryer/gotattletale/pkg"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func main() {
	dev := pkg.Device{
		Name:        "enp18s0",
		Description: "Ethernet interface",
	}
	repository := pkg.NewSqlLitePacketRepository("db.sqlite")
	go dev.Start()
	go handlePackets(repository)
	startGinServer(repository)
}

func handlePackets(repository pkg.PacketRepository) {
	packetCache := make([]pkg.AppPacket, 0)
	for v, ok := <-pkg.PacketsToCaptureQueue.ItemsChan; ok; v, ok = <-pkg.PacketsToCaptureQueue.ItemsChan {
		packetCache = append(packetCache, v)
		if len(packetCache) > 100 {
			err := repository.SavePackets(packetCache)
			if err != nil {
				log.Fatal(err)
				panic(err)
			}
			packetCache = make([]pkg.AppPacket, 0)
		}
	}
}

func startGinServer(repository pkg.PacketRepository) {

	handler := PacketHandlerImpl{
		service: PacketServiceImpl{
			storage: repository,
		},
	}
	router := gin.Default()
	router.GET("/api/v1/packets", handler.GetPackets)
	router.Run(":8080")
}

type PacketHandler interface {
	GetPackets(c *gin.Context) error
}

type PacketHandlerImpl struct {
	service PacketService
}

func (h PacketHandlerImpl) GetPackets(c *gin.Context) {
	packets, err := h.service.GetPackets()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}
	c.JSON(200, packets)
}

type PacketService interface {
	GetPackets() ([]pkg.SavedPacket, error)
}

type PacketServiceImpl struct {
	storage pkg.PacketRepository
}

func (s PacketServiceImpl) GetPackets() ([]pkg.SavedPacket, error) {
	return s.storage.GetPackets()
}
