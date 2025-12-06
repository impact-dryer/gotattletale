package service

import (
	"log"

	"github.com/impact-dryer/gotattletale/pkg"
)

func SniffAndStorePackets(repository pkg.PacketRepository) {
	go func() {
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
	}()
}
