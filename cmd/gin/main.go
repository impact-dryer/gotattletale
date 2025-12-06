package main

import (
	"github.com/gin-gonic/gin"
	"github.com/impact-dryer/gotattletale/internal/config"
	"github.com/impact-dryer/gotattletale/internal/controller"
	"github.com/impact-dryer/gotattletale/internal/service"
	"github.com/impact-dryer/gotattletale/pkg"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(config.NewAppConfig),
		fx.Provide(pkg.NewSqlLitePacketRepository),
		fx.Provide(service.NewPacketService),
		fx.Provide(controller.NewPacketController),
		fx.Invoke(service.SniffAndStorePackets),
		fx.Invoke(pkg.CreateNewDeviceAndStartSniffing),
		fx.Invoke(startGinServer),
	).Run()
}

func startGinServer(packetController controller.PacketController) {
	router := gin.Default()
	router.GET("/api/v1/packets", packetController.GetPackets)
	router.Run(":8080")
}
