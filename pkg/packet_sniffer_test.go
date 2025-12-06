package pkg

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestDeviceProcessPacketsPushesToQueue(t *testing.T) {
	originalQueue := PacketsToCaptureQueue
	defer func() { PacketsToCaptureQueue = originalQueue }()

	PacketsToCaptureQueue = PacketQueue{ItemsChan: make(chan AppPacket, 1)}

	packets := make(chan gopacket.Packet, 1)
	packets <- mustBuildPacket(t, "192.168.1.10", "192.168.1.20", 5555, 80)
	close(packets)

	dev := Device{Name: "test-device"}
	dev.processPackets(packets)

	select {
	case pkt := <-PacketsToCaptureQueue.ItemsChan:
		if pkt.DeviceID != "test-device" {
			t.Fatalf("expected device test-device, got %s", pkt.DeviceID)
		}
		if pkt.Data == nil {
			t.Fatal("expected packet data")
		}
	default:
		t.Fatal("expected packet to be pushed into queue")
	}
}

func TestDeviceStartUsesPacketStreamFactory(t *testing.T) {
	originalFactory := packetStreamFactory
	defer func() { packetStreamFactory = originalFactory }()

	originalQueue := PacketsToCaptureQueue
	defer func() { PacketsToCaptureQueue = originalQueue }()

	PacketsToCaptureQueue = PacketQueue{ItemsChan: make(chan AppPacket, 1)}

	packets := make(chan gopacket.Packet, 1)
	packets <- mustBuildPacket(t, "10.1.1.1", "10.1.1.2", 6000, 22)
	close(packets)

	cleaned := false
	packetStreamFactory = func(d *Device) (packetStream, error) {
		if d.Name != "stub" {
			return packetStream{}, errors.New("unexpected device")
		}
		return packetStream{
			packets: packets,
			cleanup: func() { cleaned = true },
		}, nil
	}

	dev := Device{Name: "stub"}
	dev.Start()

	if !cleaned {
		t.Fatal("expected cleanup to be called")
	}

	select {
	case pkt := <-PacketsToCaptureQueue.ItemsChan:
		if pkt.DeviceID != "stub" {
			t.Fatalf("expected packet with device stub, got %s", pkt.DeviceID)
		}
	default:
		t.Fatal("expected packet enqueued by Start")
	}
}

func TestDefaultPacketStreamFactory(t *testing.T) {
	originalOpen := openLiveCapture
	originalCreate := createOutputFile
	originalWriter := newPcapWriter
	originalOutput := outputfile
	originalFilter := packetfilter
	defer func() {
		openLiveCapture = originalOpen
		createOutputFile = originalCreate
		newPcapWriter = originalWriter
		outputfile = originalOutput
		packetfilter = originalFilter
	}()

	outputfile = "capture.pcap"
	packetfilter = "udp"

	writer := &stubPcapWriter{}
	newPcapWriter = func(io.Writer) pcapFileWriter { return writer }
	wc := &stubWriteCloser{}
	createOutputFile = func(string) (io.WriteCloser, error) { return wc, nil }
	packet := mustBuildPacket(t, "172.16.0.1", "172.16.0.2", 12345, 53)
	capture := &fakeCapture{
		packets: [][]byte{packet.Data()},
	}

	openLiveCapture = func(device string, snaplen int32, promisc bool, timeout time.Duration) (liveCapture, error) {
		return capture, nil
	}

	stream, err := defaultPacketStreamFactory(&Device{Name: "eth-test"})
	if err != nil {
		t.Fatalf("defaultPacketStreamFactory returned error: %v", err)
	}

	select {
	case pkt := <-stream.packets:
		if pkt == nil {
			t.Fatal("expected packet from stream")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for packet")
	}

	stream.cleanup()
	if !capture.closed {
		t.Fatal("expected capture to be closed")
	}
	if !wc.closed {
		t.Fatal("expected capture file to be closed")
	}
	if !writer.headerWritten {
		t.Fatal("expected pcap header to be written")
	}
	if capture.filter != "udp" {
		t.Fatalf("expected filter udp, got %s", capture.filter)
	}
}

func TestDefaultPacketStreamFactoryCreateFileError(t *testing.T) {
	originalCreate := createOutputFile
	originalOutput := outputfile
	defer func() {
		createOutputFile = originalCreate
		outputfile = originalOutput
	}()

	outputfile = "capture.pcap"
	createOutputFile = func(string) (io.WriteCloser, error) {
		return nil, errors.New("create failed")
	}

	if _, err := defaultPacketStreamFactory(&Device{}); err == nil {
		t.Fatal("expected error when file creation fails")
	}
}

func TestDefaultPacketStreamFactoryOpenLiveErrorRunsCleanup(t *testing.T) {
	originalCreate := createOutputFile
	originalWriter := newPcapWriter
	originalOpen := openLiveCapture
	originalOutput := outputfile
	defer func() {
		createOutputFile = originalCreate
		newPcapWriter = originalWriter
		openLiveCapture = originalOpen
		outputfile = originalOutput
	}()

	outputfile = "capture.pcap"
	wc := &stubWriteCloser{}
	createOutputFile = func(string) (io.WriteCloser, error) { return wc, nil }
	newPcapWriter = func(io.Writer) pcapFileWriter { return &stubPcapWriter{} }
	openLiveCapture = func(string, int32, bool, time.Duration) (liveCapture, error) {
		return nil, errors.New("open failed")
	}

	if _, err := defaultPacketStreamFactory(&Device{Name: "stub"}); err == nil {
		t.Fatal("expected error when openLiveCapture fails")
	}
	if !wc.closed {
		t.Fatal("expected cleanup to close the writer")
	}
}

func TestDefaultPacketStreamFactoryFilterError(t *testing.T) {
	originalOpen := openLiveCapture
	originalFilter := packetfilter
	defer func() {
		openLiveCapture = originalOpen
		packetfilter = originalFilter
	}()

	packetfilter = "tcp"
	capture := &fakeCapture{
		setFilterErr: errors.New("filter failed"),
	}
	openLiveCapture = func(string, int32, bool, time.Duration) (liveCapture, error) {
		return capture, nil
	}

	if _, err := defaultPacketStreamFactory(&Device{}); err == nil {
		t.Fatal("expected error when filter setup fails")
	}
	if !capture.closed {
		t.Fatal("expected capture closed after filter failure")
	}
}

func TestRunCleanups(t *testing.T) {
	var order []int
	runCleanups([]func(){
		func() { order = append(order, 1) },
		func() { order = append(order, 2) },
	})

	if len(order) != 2 || order[0] != 2 || order[1] != 1 {
		t.Fatalf("expected cleanup order [2 1], got %v", order)
	}
}

type fakeCapture struct {
	packets      [][]byte
	filter       string
	closed       bool
	setFilterErr error
}

func (f *fakeCapture) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	if len(f.packets) == 0 {
		return nil, gopacket.CaptureInfo{}, io.EOF
	}
	data := f.packets[0]
	f.packets = f.packets[1:]
	return data, gopacket.CaptureInfo{Timestamp: time.Now()}, nil
}

func (f *fakeCapture) Close() {
	f.closed = true
}

func (f *fakeCapture) LinkType() layers.LinkType {
	return layers.LinkTypeEthernet
}

func (f *fakeCapture) SetBPFFilter(filter string) error {
	f.filter = filter
	return f.setFilterErr
}

type stubWriteCloser struct {
	bytes.Buffer
	closed bool
}

func (s *stubWriteCloser) Close() error {
	s.closed = true
	return nil
}

type stubPcapWriter struct {
	headerWritten bool
	err           error
}

func (s *stubPcapWriter) WriteFileHeader(uint32, layers.LinkType) error {
	s.headerWritten = true
	return s.err
}

func mustBuildPacket(t *testing.T, srcIP, dstIP string, srcPort, dstPort int) gopacket.Packet {
	t.Helper()

	ipLayer := &layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
		SrcIP:    parseIP(srcIP),
		DstIP:    parseIP(dstIP),
	}

	tcpLayer := &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		Seq:     1,
		SYN:     true,
	}
	tcpLayer.SetNetworkLayerForChecksum(ipLayer)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	if err := gopacket.SerializeLayers(buf, opts, ipLayer, tcpLayer); err != nil {
		t.Fatalf("failed to serialize packet: %v", err)
	}

	return gopacket.NewPacket(buf.Bytes(), layers.LayerTypeIPv4, gopacket.Default)
}

func parseIP(ip string) []byte {
	var result []byte
	var num byte
	for i := 0; i < len(ip); i++ {
		if ip[i] == '.' {
			result = append(result, num)
			num = 0
		} else {
			num = num*10 + (ip[i] - '0')
		}
	}
	result = append(result, num)
	return result
}
