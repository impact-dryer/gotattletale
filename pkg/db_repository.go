package pkg

import (
	"database/sql"
	"log"
	"os"
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
		INSERT INTO packets (data, created_at, updated_at, device_id)
		VALUES (?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, packet.Data.Dump(), packet.CreatedAt, packet.UpdatedAt, packet.DeviceID)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return nil
}

func (r *SqlLitePacketRepository) SavePackets(packets []AppPacket) error {
	query := `
		INSERT INTO packets (data, created_at, updated_at, device_id)
		VALUES (?, ?, ?, ?)
	`
	for _, packet := range packets {
		_, err := r.db.Exec(query, packet.Data.Dump(), packet.CreatedAt, packet.UpdatedAt, packet.DeviceID)
		if err != nil {
			log.Fatal(err)
			panic(err)
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
	schema, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return string(schema), nil
}
func InitSchema(db *sql.DB) error {
	schema, err := ReadSchemaFile("schema.sql")
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return nil
}

func NewSqlLitePacketRepository(db *sql.DB) *SqlLitePacketRepository {
	return &SqlLitePacketRepository{db: db}
}
