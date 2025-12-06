package pkg

import (
	"log"
	"strconv"
	"time"

	"github.com/google/gopacket"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AppPacket struct {
	ID        string
	Data      gopacket.Packet
	CreatedAt time.Time
	UpdatedAt time.Time
	DeviceID  string
}

type SavedPacket struct {
	ID              uint      `gorm:"primaryKey"`
	SourceIP        string    `gorm:"not null"`
	DestinationIP   string    `gorm:"not null"`
	SourcePort      int       `gorm:"not null"`
	DestinationPort int       `gorm:"not null"`
	Protocol        string    `gorm:"not null"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
	DeviceID        string    `gorm:"not null"`
}

type PacketRepository interface {
	SavePacket(packet AppPacket) error
	SavePackets(packets []AppPacket) error
	GetPackets() ([]SavedPacket, error)
	GetPacket(packetID string) (SavedPacket, error)
	DeletePacket(packetID string) error
	UpdatePacket(packet AppPacket) error
}

type SqlLitePacketRepository struct {
	db *gorm.DB
}

func (r *SqlLitePacketRepository) SavePacket(packet AppPacket) error {
	savedPacket, err := mapPacketToSavedPacket(packet)
	if err != nil {
		return err
	}
	r.db.Create(savedPacket)
	return nil
}

func mapPacketToSavedPacket(packet AppPacket) (*SavedPacket, error) {
	sourceIP := packet.Data.NetworkLayer().NetworkFlow().Src().String()
	destinationIP := packet.Data.NetworkLayer().NetworkFlow().Dst().String()
	sourcePort := packet.Data.TransportLayer().TransportFlow().Src().String()
	destinationPort := packet.Data.TransportLayer().TransportFlow().Dst().String()
	protocol := packet.Data.TransportLayer().LayerType().String()
	sourcePortInt, err := strconv.Atoi(sourcePort)
	if err != nil {
		return nil, err
	}
	destinationPortInt, err := strconv.Atoi(destinationPort)
	if err != nil {
		return nil, err
	}
	savedPacket := &SavedPacket{
		SourceIP:        sourceIP,
		DestinationIP:   destinationIP,
		SourcePort:      sourcePortInt,
		DestinationPort: destinationPortInt,
		Protocol:        protocol,
		CreatedAt:       packet.CreatedAt,
		UpdatedAt:       packet.UpdatedAt,
		DeviceID:        packet.DeviceID,
	}
	return savedPacket, nil
}

func (r *SqlLitePacketRepository) SavePackets(packets []AppPacket) error {
	mapedPackets := make([]*SavedPacket, len(packets))
	for i, packet := range packets {
		packet, err := mapPacketToSavedPacket(packet)
		if err != nil {
			return err
		}
		mapedPackets[i] = packet
	}
	r.db.Create(mapedPackets)
	return nil
}

func (r *SqlLitePacketRepository) GetPackets() ([]SavedPacket, error) {
	packets := make([]SavedPacket, 100)
	result := r.db.Find(&packets)
	if result.Error != nil {
		return nil, result.Error
	}
	return packets, nil
}

func (r *SqlLitePacketRepository) GetPacket(packetID string) (SavedPacket, error) {
	return SavedPacket{}, nil
}

func (r *SqlLitePacketRepository) DeletePacket(packetID string) error {
	return nil
}

func (r *SqlLitePacketRepository) UpdatePacket(packet AppPacket) error {
	return nil
}

func NewSqlLitePacketRepository(dbName string) PacketRepository {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
		return nil
	}
	// Migrate the schema
	db.AutoMigrate(&SavedPacket{})

	return &SqlLitePacketRepository{db: db}
}
