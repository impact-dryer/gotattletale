package main

import (
	"database/sql"
	"log"

	"github.com/impact-dryer/gotattletale/pkg"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func main() {
	tattletaleDb, err := sql.Open("sqlite3", "tattletale.db")
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	defer tattletaleDb.Close()
	pkg.InitSchema(tattletaleDb)
	dev := pkg.Device{
		Name:        "enp18s0",
		Description: "Ethernet interface",
	}
	repository := pkg.NewSqlLitePacketRepository(tattletaleDb)
	go dev.Start()
	for v, ok := <-pkg.PacketsToCaptureQueue.ItemsChan; ok; v, ok = <-pkg.PacketsToCaptureQueue.ItemsChan {
		log.Println("Saving packet to database", v)
		log.Println("Packet data", v.Data.Dump())
		err = repository.SavePacket(v)
		if err != nil {
			log.Fatal(err)
			panic(err)
		}
	}
}
