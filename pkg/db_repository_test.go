package pkg

import (
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *SqlLitePacketRepository {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	db.AutoMigrate(&SavedPacket{})
	return &SqlLitePacketRepository{db: db}
}

func createTestPacket(srcIP, dstIP string, srcPort, dstPort int) gopacket.Packet {
	// Create a mock TCP/IP packet
	ipLayer := &layers.IPv4{
		SrcIP:    []byte{192, 168, 1, 1},
		DstIP:    []byte{192, 168, 1, 2},
		Protocol: layers.IPProtocolTCP,
	}
	if srcIP == "10.0.0.1" {
		ipLayer.SrcIP = []byte{10, 0, 0, 1}
	}
	if dstIP == "10.0.0.2" {
		ipLayer.DstIP = []byte{10, 0, 0, 2}
	}

	tcpLayer := &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
	}
	tcpLayer.SetNetworkLayerForChecksum(ipLayer)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	gopacket.SerializeLayers(buf, opts, ipLayer, tcpLayer)

	return gopacket.NewPacket(buf.Bytes(), layers.LayerTypeIPv4, gopacket.Default)
}

func TestSavePacket(t *testing.T) {
	repo := setupTestDB(t)

	packet := AppPacket{
		ID:        "test-id-1",
		Data:      createTestPacket("192.168.1.1", "192.168.1.2", 8080, 443),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeviceID:  "eth0",
	}

	err := repo.SavePacket(packet)
	if err != nil {
		t.Fatalf("failed to save packet: %v", err)
	}

	// Verify the packet was saved
	var savedPacket SavedPacket
	result := repo.db.First(&savedPacket)
	if result.Error != nil {
		t.Fatalf("failed to retrieve saved packet: %v", result.Error)
	}

	if savedPacket.DeviceID != "eth0" {
		t.Errorf("expected DeviceID 'eth0', got '%s'", savedPacket.DeviceID)
	}
	if savedPacket.SourcePort != 8080 {
		t.Errorf("expected SourcePort 8080, got %d", savedPacket.SourcePort)
	}
	if savedPacket.DestinationPort != 443 {
		t.Errorf("expected DestinationPort 443, got %d", savedPacket.DestinationPort)
	}
	if savedPacket.Protocol != "TCP" {
		t.Errorf("expected Protocol 'TCP', got '%s'", savedPacket.Protocol)
	}
}

func TestSavePackets(t *testing.T) {
	repo := setupTestDB(t)

	packets := []AppPacket{
		{
			ID:        "test-id-1",
			Data:      createTestPacket("192.168.1.1", "192.168.1.2", 8080, 443),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeviceID:  "eth0",
		},
		{
			ID:        "test-id-2",
			Data:      createTestPacket("10.0.0.1", "10.0.0.2", 9000, 80),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeviceID:  "eth1",
		},
	}

	err := repo.SavePackets(packets)
	if err != nil {
		t.Fatalf("failed to save packets: %v", err)
	}

	// Verify packets were saved
	var count int64
	repo.db.Model(&SavedPacket{}).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 packets saved, got %d", count)
	}
}

func TestGetPackets(t *testing.T) {
	repo := setupTestDB(t)

	// Insert test data directly
	now := time.Now()
	testPackets := []SavedPacket{
		{
			SourceIP:        "192.168.1.1",
			DestinationIP:   "192.168.1.2",
			SourcePort:      8080,
			DestinationPort: 443,
			Protocol:        "TCP",
			CreatedAt:       now.Add(-time.Hour),
			UpdatedAt:       now,
			DeviceID:        "eth0",
		},
		{
			SourceIP:        "10.0.0.1",
			DestinationIP:   "10.0.0.2",
			SourcePort:      9000,
			DestinationPort: 80,
			Protocol:        "TCP",
			CreatedAt:       now,
			UpdatedAt:       now,
			DeviceID:        "eth1",
		},
	}
	repo.db.Create(&testPackets)

	// Test GetPackets with default sort
	packets, err := repo.GetPackets(10, "")
	if err != nil {
		t.Fatalf("failed to get packets: %v", err)
	}

	if len(packets) < 2 {
		t.Fatalf("expected at least 2 packets, got %d", len(packets))
	}

	// Verify ordering (should be desc by created_at, so most recent first)
	if packets[0].DeviceID != "eth1" {
		t.Errorf("expected first packet DeviceID 'eth1', got '%s'", packets[0].DeviceID)
	}
}

func TestGetPacketsWithLimit(t *testing.T) {
	repo := setupTestDB(t)

	// Insert 5 test packets
	now := time.Now()
	for i := 0; i < 5; i++ {
		packet := SavedPacket{
			SourceIP:        "192.168.1.1",
			DestinationIP:   "192.168.1.2",
			SourcePort:      8080 + i,
			DestinationPort: 443,
			Protocol:        "TCP",
			CreatedAt:       now.Add(time.Duration(i) * time.Minute),
			UpdatedAt:       now,
			DeviceID:        "eth0",
		}
		repo.db.Create(&packet)
	}

	// Test with limit of 3
	packets, err := repo.GetPackets(3, "")
	if err != nil {
		t.Fatalf("failed to get packets: %v", err)
	}

	if len(packets) != 3 {
		t.Errorf("expected 3 packets with limit, got %d", len(packets))
	}
}

func TestGetPacketsWithCustomSort(t *testing.T) {
	repo := setupTestDB(t)

	// Insert test data
	now := time.Now()
	testPackets := []SavedPacket{
		{
			SourceIP:        "192.168.1.1",
			DestinationIP:   "192.168.1.2",
			SourcePort:      1000,
			DestinationPort: 443,
			Protocol:        "TCP",
			CreatedAt:       now,
			UpdatedAt:       now,
			DeviceID:        "eth0",
		},
		{
			SourceIP:        "10.0.0.1",
			DestinationIP:   "10.0.0.2",
			SourcePort:      9000,
			DestinationPort: 80,
			Protocol:        "TCP",
			CreatedAt:       now,
			UpdatedAt:       now,
			DeviceID:        "eth1",
		},
	}
	repo.db.Create(&testPackets)

	// Test GetPackets with source_port sort
	packets, err := repo.GetPackets(10, "source_port")
	if err != nil {
		t.Fatalf("failed to get packets: %v", err)
	}

	// Should be sorted desc by source_port
	if packets[0].SourcePort != 9000 {
		t.Errorf("expected first packet SourcePort 9000, got %d", packets[0].SourcePort)
	}
}

func TestGetPacket(t *testing.T) {
	repo := setupTestDB(t)

	// GetPacket is not implemented, just verify it returns empty
	packet, err := repo.GetPacket("any-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return empty SavedPacket (not implemented)
	if packet.ID != 0 {
		t.Errorf("expected empty packet, got ID %d", packet.ID)
	}
}

func TestDeletePacket(t *testing.T) {
	repo := setupTestDB(t)

	// DeletePacket is not implemented, just verify no error
	err := repo.DeletePacket("any-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdatePacket(t *testing.T) {
	repo := setupTestDB(t)

	// UpdatePacket is not implemented, just verify no error
	err := repo.UpdatePacket(AppPacket{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMapPacketToSavedPacket(t *testing.T) {
	now := time.Now()
	packet := AppPacket{
		ID:        "test-id",
		Data:      createTestPacket("192.168.1.1", "192.168.1.2", 8080, 443),
		CreatedAt: now,
		UpdatedAt: now,
		DeviceID:  "eth0",
	}

	savedPacket, err := mapPacketToSavedPacket(packet)
	if err != nil {
		t.Fatalf("failed to map packet: %v", err)
	}

	if savedPacket.SourcePort != 8080 {
		t.Errorf("expected SourcePort 8080, got %d", savedPacket.SourcePort)
	}
	if savedPacket.DestinationPort != 443 {
		t.Errorf("expected DestinationPort 443, got %d", savedPacket.DestinationPort)
	}
	if savedPacket.Protocol != "TCP" {
		t.Errorf("expected Protocol 'TCP', got '%s'", savedPacket.Protocol)
	}
	if savedPacket.DeviceID != "eth0" {
		t.Errorf("expected DeviceID 'eth0', got '%s'", savedPacket.DeviceID)
	}
	if !savedPacket.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt %v, got %v", now, savedPacket.CreatedAt)
	}
}
