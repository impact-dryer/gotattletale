package pkg

import (
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/impact-dryer/gotattletale/internal/config"
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
	ItemsChan: make(chan AppPacket),
}

type packetStream struct {
	packets <-chan gopacket.Packet
	cleanup func()
}

var packetStreamFactory = defaultPacketStreamFactory

type liveCapture interface {
	gopacket.PacketDataSource
	Close()
	LinkType() layers.LinkType
	SetBPFFilter(string) error
}

type pcapFileWriter interface {
	WriteFileHeader(snaplen uint32, linkType layers.LinkType) error
}

var (
	openLiveCapture = func(device string, snaplen int32, promisc bool, timeout time.Duration) (liveCapture, error) {
		return pcap.OpenLive(device, snaplen, promisc, timeout)
	}
	createOutputFile = func(name string) (io.WriteCloser, error) {
		return os.Create(name)
	}
	newPcapWriter = func(w io.Writer) pcapFileWriter {
		return pcapgo.NewWriter(w)
	}
)

// Start capturing packets
func (d *Device) Start() {
	stream, err := packetStreamFactory(d)
	if err != nil {
		log.Fatal(err)
	}
	if stream.cleanup != nil {
		defer stream.cleanup()
	}

	d.processPackets(stream.packets)
}

func (d *Device) processPackets(packets <-chan gopacket.Packet) {
	for packet := range packets {
		log.Println("Pushing packet to queue")
		PacketsToCaptureQueue.Push(AppPacket{
			Data:      packet,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeviceID:  d.Name,
		})
	}
}

func defaultPacketStreamFactory(d *Device) (packetStream, error) {
	var cleanups []func()
	if outputfile != "" {
		f, err := createOutputFile(outputfile)
		if err != nil {
			return packetStream{}, err
		}
		writer := newPcapWriter(f)
		if err := writer.WriteFileHeader(uint32(SNAPSHOTLENGTH), layers.LinkTypeEthernet); err != nil {
			f.Close()
			return packetStream{}, err
		}
		cleanups = append(cleanups, func() { f.Close() })
	}

	handler, err := openLiveCapture(d.Name, SNAPSHOTLENGTH, PROMISCUOUS, TIMEOUT)
	if err != nil {
		runCleanups(cleanups)
		return packetStream{}, err
	}
	cleanups = append(cleanups, handler.Close)

	if packetfilter != "" {
		if err := handler.SetBPFFilter(packetfilter); err != nil {
			runCleanups(cleanups)
			return packetStream{}, err
		}
	}

	source := gopacket.NewPacketSource(handler, handler.LinkType())
	return packetStream{
		packets: source.Packets(),
		cleanup: func() { runCleanups(cleanups) },
	}, nil
}

func runCleanups(cleanups []func()) {
	for i := len(cleanups) - 1; i >= 0; i-- {
		if cleanups[i] != nil {
			cleanups[i]()
		}
	}
}

func CreateNewDeviceAndStartSniffing(appconfig *config.AppConfig) {
	device := Device{
		Name: appconfig.DeviceName,
	}
	go device.Start()
}
