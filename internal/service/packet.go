package service

import "github.com/impact-dryer/gotattletale/pkg"

type PacketService interface {
	GetPackets() ([]pkg.SavedPacket, error)
}

type PacketServiceImpl struct {
	Storage pkg.PacketRepository
}

func (s PacketServiceImpl) GetPackets() ([]pkg.SavedPacket, error) {
	return s.Storage.GetPackets()
}

func NewPacketService(storage pkg.PacketRepository) PacketService {
	return &PacketServiceImpl{Storage: storage}
}
