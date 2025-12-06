package service

import (
	"errors"
	"testing"
	"time"

	"github.com/impact-dryer/gotattletale/pkg"
)

// MockPacketRepository is a mock implementation of PacketRepository
type MockPacketRepository struct {
	packets         []pkg.SavedPacket
	getPacketsErr   error
	savePacketErr   error
	savePacketsErr  error
	calledWithLimit int
	calledWithSort  string
	savedPacket     pkg.AppPacket
	savedPackets    []pkg.AppPacket
}

func (m *MockPacketRepository) SavePacket(packet pkg.AppPacket) error {
	m.savedPacket = packet
	return m.savePacketErr
}

func (m *MockPacketRepository) SavePackets(packets []pkg.AppPacket) error {
	m.savedPackets = packets
	return m.savePacketsErr
}

func (m *MockPacketRepository) GetPackets(limit int, sort string) ([]pkg.SavedPacket, error) {
	m.calledWithLimit = limit
	m.calledWithSort = sort
	return m.packets, m.getPacketsErr
}

func (m *MockPacketRepository) GetPacket(packetID string) (pkg.SavedPacket, error) {
	return pkg.SavedPacket{}, nil
}

func (m *MockPacketRepository) DeletePacket(packetID string) error {
	return nil
}

func (m *MockPacketRepository) UpdatePacket(packet pkg.AppPacket) error {
	return nil
}

func TestPacketService_GetPackets_Success(t *testing.T) {
	// Arrange
	now := time.Now()
	expectedPackets := []pkg.SavedPacket{
		{
			ID:              1,
			SourceIP:        "192.168.1.1",
			DestinationIP:   "192.168.1.2",
			SourcePort:      8080,
			DestinationPort: 443,
			Protocol:        "TCP",
			CreatedAt:       now,
			UpdatedAt:       now,
			DeviceID:        "eth0",
		},
		{
			ID:              2,
			SourceIP:        "10.0.0.1",
			DestinationIP:   "10.0.0.2",
			SourcePort:      3000,
			DestinationPort: 80,
			Protocol:        "UDP",
			CreatedAt:       now,
			UpdatedAt:       now,
			DeviceID:        "eth1",
		},
	}

	mockRepo := &MockPacketRepository{
		packets:       expectedPackets,
		getPacketsErr: nil,
	}

	service := NewPacketService(mockRepo)

	// Act
	packets, err := service.GetPackets(50, "source_ip")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(packets) != len(expectedPackets) {
		t.Errorf("expected %d packets, got %d", len(expectedPackets), len(packets))
	}

	if mockRepo.calledWithLimit != 50 {
		t.Errorf("expected limit 50, got %d", mockRepo.calledWithLimit)
	}

	if mockRepo.calledWithSort != "source_ip" {
		t.Errorf("expected sort 'source_ip', got '%s'", mockRepo.calledWithSort)
	}
}

func TestPacketService_GetPackets_RepositoryError(t *testing.T) {
	// Arrange
	expectedError := errors.New("database connection failed")
	mockRepo := &MockPacketRepository{
		packets:       nil,
		getPacketsErr: expectedError,
	}

	service := NewPacketService(mockRepo)

	// Act
	packets, err := service.GetPackets(100, "created_at")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%s', got '%s'", expectedError.Error(), err.Error())
	}

	if packets != nil {
		t.Errorf("expected nil packets on error, got %v", packets)
	}
}

func TestPacketService_GetPackets_EmptyResult(t *testing.T) {
	// Arrange
	mockRepo := &MockPacketRepository{
		packets:       []pkg.SavedPacket{},
		getPacketsErr: nil,
	}

	service := NewPacketService(mockRepo)

	// Act
	packets, err := service.GetPackets(100, "")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(packets) != 0 {
		t.Errorf("expected 0 packets, got %d", len(packets))
	}
}

func TestPacketService_GetPackets_DefaultSort(t *testing.T) {
	// Arrange
	mockRepo := &MockPacketRepository{
		packets:       []pkg.SavedPacket{},
		getPacketsErr: nil,
	}

	service := NewPacketService(mockRepo)

	// Act
	_, err := service.GetPackets(10, "")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// The service passes through whatever sort value it receives
	// The repository should handle the default
	if mockRepo.calledWithSort != "" {
		t.Errorf("expected empty sort to be passed through, got '%s'", mockRepo.calledWithSort)
	}
}

func TestNewPacketService(t *testing.T) {
	// Arrange
	mockRepo := &MockPacketRepository{}

	// Act
	service := NewPacketService(mockRepo)

	// Assert
	if service == nil {
		t.Fatal("expected non-nil service")
	}

	impl, ok := service.(*PacketServiceImpl)
	if !ok {
		t.Fatal("expected service to be *PacketServiceImpl")
	}

	if impl.Storage != mockRepo {
		t.Error("expected storage to be the mock repository")
	}
}

func TestPacketServiceImpl_GetPackets_PassesParametersCorrectly(t *testing.T) {
	testCases := []struct {
		name          string
		limit         int
		sort          string
		expectedLimit int
		expectedSort  string
	}{
		{
			name:          "standard parameters",
			limit:         100,
			sort:          "created_at",
			expectedLimit: 100,
			expectedSort:  "created_at",
		},
		{
			name:          "custom limit and sort",
			limit:         25,
			sort:          "destination_ip",
			expectedLimit: 25,
			expectedSort:  "destination_ip",
		},
		{
			name:          "zero limit",
			limit:         0,
			sort:          "source_port",
			expectedLimit: 0,
			expectedSort:  "source_port",
		},
		{
			name:          "negative limit",
			limit:         -1,
			sort:          "protocol",
			expectedLimit: -1,
			expectedSort:  "protocol",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &MockPacketRepository{
				packets: []pkg.SavedPacket{},
			}

			service := &PacketServiceImpl{Storage: mockRepo}

			_, _ = service.GetPackets(tc.limit, tc.sort)

			if mockRepo.calledWithLimit != tc.expectedLimit {
				t.Errorf("expected limit %d, got %d", tc.expectedLimit, mockRepo.calledWithLimit)
			}

			if mockRepo.calledWithSort != tc.expectedSort {
				t.Errorf("expected sort '%s', got '%s'", tc.expectedSort, mockRepo.calledWithSort)
			}
		})
	}
}
