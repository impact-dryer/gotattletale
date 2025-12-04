package pkg

import (
	"database/sql"
	"errors"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	_ "github.com/mattn/go-sqlite3"
)

func TestSqlLitePacketRepository_SavePacket(t *testing.T) {
	db := newTestDB(t)
	repo := NewSqlLitePacketRepository(db)

	appPacket := AppPacket{
		Data:      mustBuildPacket(t, "10.0.0.1", "10.0.0.2", 1234, 80),
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
		DeviceID:  "eth0",
	}

	if err := repo.SavePacket(appPacket); err != nil {
		t.Fatalf("SavePacket returned error: %v", err)
	}

	var (
		sourceIP        string
		destinationIP   string
		sourcePort      string
		destinationPort string
		protocol        string
		deviceID        string
	)

	row := db.QueryRow(`
		SELECT source_ip, destination_ip, source_port, destination_port, protocol, device_id
		FROM packets
	`)
	if err := row.Scan(&sourceIP, &destinationIP, &sourcePort, &destinationPort, &protocol, &deviceID); err != nil {
		t.Fatalf("failed to scan row: %v", err)
	}

	if sourceIP != "10.0.0.1" || destinationIP != "10.0.0.2" {
		t.Fatalf("unexpected IPs: %s -> %s", sourceIP, destinationIP)
	}
	if sourcePort != "1234" || destinationPort != "80" {
		t.Fatalf("unexpected ports: %s -> %s", sourcePort, destinationPort)
	}
	if protocol != "TCP" {
		t.Fatalf("unexpected protocol: %s", protocol)
	}
	if deviceID != "eth0" {
		t.Fatalf("unexpected device id: %s", deviceID)
	}
}

func TestSqlLitePacketRepository_SavePackets(t *testing.T) {
	db := newTestDB(t)
	repo := NewSqlLitePacketRepository(db)

	appPackets := []AppPacket{
		{
			Data:      mustBuildPacket(t, "10.10.0.1", "10.10.0.2", 1000, 2000),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			DeviceID:  "eth0",
		},
		{
			Data:      mustBuildPacket(t, "10.10.0.3", "10.10.0.4", 3000, 4000),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			DeviceID:  "eth1",
		},
	}

	if err := repo.SavePackets(appPackets); err != nil {
		t.Fatalf("SavePackets returned error: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM packets`).Scan(&count); err != nil {
		t.Fatalf("failed to count packets: %v", err)
	}
	if count != len(appPackets) {
		t.Fatalf("expected %d packets, got %d", len(appPackets), count)
	}
}

func TestSqlLitePacketRepository_SavePacketError(t *testing.T) {
	db := newTestDB(t)
	repo := NewSqlLitePacketRepository(db)
	packet := AppPacket{
		Data:      mustBuildPacket(t, "10.0.0.1", "10.0.0.2", 1000, 2000),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeviceID:  "eth0",
	}

	db.Close()
	if err := repo.SavePacket(packet); err == nil {
		t.Fatal("expected error when saving packet on closed db")
	}
}

func TestSqlLitePacketRepository_SavePacketsError(t *testing.T) {
	db := newTestDB(t)
	repo := NewSqlLitePacketRepository(db)
	packets := []AppPacket{
		{
			Data:      mustBuildPacket(t, "1.1.1.1", "2.2.2.2", 1111, 80),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeviceID:  "eth0",
		},
	}
	db.Close()
	if err := repo.SavePackets(packets); err == nil {
		t.Fatal("expected error when bulk saving packets on closed db")
	}
}

func TestInitSchemaCreatesTables(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	var tableName string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='packets'`).Scan(&tableName)
	if err != nil {
		t.Fatalf("packets table not created: %v", err)
	}
}

func TestInitSchemaPropagatesReadError(t *testing.T) {
	originalReader := schemaFileReader
	defer func() { schemaFileReader = originalReader }()

	schemaFileReader = func(string) (string, error) {
		return "", errors.New("read failure")
	}

	db := openTestDB(t)
	if err := InitSchema(db); err == nil {
		t.Fatal("expected error when schema cannot be read")
	}
}

func TestInitSchemaPropagatesExecError(t *testing.T) {
	originalReader := schemaFileReader
	defer func() { schemaFileReader = originalReader }()

	schemaFileReader = func(string) (string, error) {
		return "INVALID SQL;", nil
	}
	db := openTestDB(t)
	if err := InitSchema(db); err == nil {
		t.Fatal("expected error when schema execution fails")
	}
}

func TestReadSchemaFile(t *testing.T) {
	content, err := ReadSchemaFile(filepath.Join("schema.sql"))
	if err != nil {
		t.Fatalf("ReadSchemaFile returned error: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("expected schema content, got empty string")
	}
}

func TestReadSchemaFileMissing(t *testing.T) {
	if _, err := ReadSchemaFile("does_not_exist.sql"); err == nil {
		t.Fatal("expected error for missing schema file")
	}
}

func TestSqlLitePacketRepository_UnimplementedMethods(t *testing.T) {
	repo := &SqlLitePacketRepository{}

	if packets, err := repo.GetPackets(); err != nil || packets != nil {
		t.Fatalf("GetPackets expected nil,nil got %v,%v", packets, err)
	}

	if packet, err := repo.GetPacket("id"); err != nil || (packet != AppPacket{}) {
		t.Fatalf("GetPacket expected zero packet,nil got %v,%v", packet, err)
	}

	if err := repo.DeletePacket("id"); err != nil {
		t.Fatalf("DeletePacket expected nil, got %v", err)
	}

	if err := repo.UpdatePacket(AppPacket{}); err != nil {
		t.Fatalf("UpdatePacket expected nil, got %v", err)
	}
}

func newTestDB(t *testing.T) *sql.DB {
	db := openTestDB(t)
	t.Cleanup(func() {
		db.Close()
	})

	if err := InitSchema(db); err != nil {
		t.Fatalf("failed to init schema: %v", err)
	}
	return db
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	return db
}

func mustBuildPacket(t *testing.T, srcIP, dstIP string, srcPort, dstPort int) gopacket.Packet {
	t.Helper()

	ipSrc := net.ParseIP(srcIP)
	ipDst := net.ParseIP(dstIP)
	if ipSrc == nil || ipDst == nil {
		t.Fatalf("invalid IP provided: %s -> %s", srcIP, dstIP)
	}

	eth := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
		DstMAC:       net.HardwareAddr{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa},
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip := &layers.IPv4{
		SrcIP:    ipSrc,
		DstIP:    ipDst,
		Protocol: layers.IPProtocolTCP,
	}
	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		SYN:     true,
	}
	tcp.SetNetworkLayerForChecksum(ip)

	payload := gopacket.Payload([]byte("payload"))
	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	if err := gopacket.SerializeLayers(buffer, opts, eth, ip, tcp, payload); err != nil {
		t.Fatalf("failed to serialize packet: %v", err)
	}

	return gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
}
