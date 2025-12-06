package service

import "github.com/impact-dryer/gotattletale/pkg"

type PacketService interface {
	GetPackets(limit int, sort string) ([]pkg.SavedPacket, error)
}

type PacketServiceImpl struct {
	Storage pkg.PacketRepository
}

func (s PacketServiceImpl) GetPackets(limit int, sort string) ([]pkg.SavedPacket, error) {
	return s.Storage.GetPackets(limit, sort)
}

func NewPacketService(storage pkg.PacketRepository) PacketService {
	return &PacketServiceImpl{Storage: storage}
}
