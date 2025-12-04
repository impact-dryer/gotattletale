package pkg

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/google/gopacket"
)

type AppPacket struct {
	ID        string
	Data      gopacket.Packet
	CreatedAt time.Time
	UpdatedAt time.Time
	DeviceID  string
}

type PacketRepository interface {
	SavePacket(packet AppPacket) error
	SavePackets(packets []AppPacket) error
	GetPackets() ([]AppPacket, error)
	GetPacket(packetID string) (AppPacket, error)
	DeletePacket(packetID string) error
	UpdatePacket(packet AppPacket) error
}

type SqlLitePacketRepository struct {
	db *sql.DB
}

func (r *SqlLitePacketRepository) SavePacket(packet AppPacket) error {
	query := `
	INSERT INTO packets (source_ip, destination_ip, source_port, destination_port, protocol, created_at, updated_at, device_id)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	sourceIP := packet.Data.NetworkLayer().NetworkFlow().Src().String()
	destinationIP := packet.Data.NetworkLayer().NetworkFlow().Dst().String()
	sourcePort := packet.Data.TransportLayer().TransportFlow().Src().String()
	destinationPort := packet.Data.TransportLayer().TransportFlow().Dst().String()
	protocol := packet.Data.TransportLayer().LayerType().String()
	_, err := r.db.Exec(query, sourceIP, destinationIP, sourcePort, destinationPort, protocol, packet.CreatedAt, packet.UpdatedAt, packet.DeviceID)
	return err
}

func (r *SqlLitePacketRepository) SavePackets(packets []AppPacket) error {
	for _, packet := range packets {
		if err := r.SavePacket(packet); err != nil {
			return err
		}
	}
	return nil
}

func (r *SqlLitePacketRepository) GetPackets() ([]AppPacket, error) {
	return nil, nil
}

func (r *SqlLitePacketRepository) GetPacket(packetID string) (AppPacket, error) {
	return AppPacket{}, nil
}

func (r *SqlLitePacketRepository) DeletePacket(packetID string) error {
	return nil
}

func (r *SqlLitePacketRepository) UpdatePacket(packet AppPacket) error {
	return nil
}

// Initialize the schema for the database from schema.sql file
func ReadSchemaFile(filename string) (string, error) {
	pathsToTry := []string{filename}
	if !filepath.IsAbs(filename) {
		pathsToTry = append(pathsToTry, filepath.Join("..", filename))
	}

	var lastErr error
	for _, path := range pathsToTry {
		schema, err := os.ReadFile(path)
		if err == nil {
			return string(schema), nil
		}
		lastErr = err
	}
	return "", lastErr
}
func InitSchema(db *sql.DB) error {
	if db == nil {
		return errors.New("nil database handle")
	}
	schema, err := schemaFileReader("schema.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(schema)
	return err
}

func NewSqlLitePacketRepository(db *sql.DB) *SqlLitePacketRepository {
	return &SqlLitePacketRepository{db: db}
}

var schemaFileReader = ReadSchemaFile
