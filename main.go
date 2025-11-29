package main

import (
	"database/sql"
	"log"
	"time"

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
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("Popping packet from queue")
				packets, err := pkg.PacketsToCaptureQueue.PopMultiple(100)
				if err != nil {
					log.Println("Error popping packets from queue", err)
				} else {
					log.Println("Saving packet to database")
					err = repository.SavePackets(packets)
					if err != nil {
						log.Fatal(err)
					}
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	dev.Start()
}
