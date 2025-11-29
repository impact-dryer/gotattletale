package pkg

import (
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
)

type Devices struct {
	devices []Device
}

type Device struct {
	ID          int
	Name        string
	Description string
	MAC         net.HardwareAddr
	Addresses   []Addrs
}

type Addrs struct {
	IP      net.IP
	Netmask net.IPMask
}

const (
	SNAPSHOTLENGTH = 65535
	PROMISCUOUS    = true
	TIMEOUT        = 30 * time.Second
)

var outputfile = "output.pcap"
var packetfilter = "tcp"

var PacketsToCaptureQueue = PacketQueue{
	Items: make([]AppPacket, 0),
	Mutex: sync.Mutex{},
}

// Start capturing packets
func (d *Device) Start() {
	// If user wants to save the data to a file
	var w *pcapgo.Writer
	if outputfile != "" {
		// Open output pcap file and write header
		f, _ := os.Create(outputfile)
		w = pcapgo.NewWriter(f)
		w.WriteFileHeader(uint32(SNAPSHOTLENGTH), layers.LinkTypeEthernet)
		defer f.Close()
	}

	// Open the device for capturing
	handler, err := pcap.OpenLive(d.Name, SNAPSHOTLENGTH, PROMISCUOUS, TIMEOUT)
	if err != nil {
		log.Fatal(err)
	}
	defer handler.Close()

	// Set filter if one was provided
	if packetfilter != "" {
		err := handler.SetBPFFilter(packetfilter)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Start processing packets
	source := gopacket.NewPacketSource(handler, handler.LinkType())

	for packet := range source.Packets() {
		log.Println("Pushing packet to queue")
		PacketsToCaptureQueue.Push(AppPacket{
			Data:      packet,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeviceID:  d.Name,
		})
	}

}
