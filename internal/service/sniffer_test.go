package service

import (
	"sync"
	"testing"
	"time"

	"github.com/impact-dryer/gotattletale/pkg"
)

// TestMockPacketRepository for sniffer tests
type TestMockPacketRepository struct {
	mu             sync.Mutex
	savedPackets   [][]pkg.AppPacket
	savePacketsErr error
	saveCalled     chan struct{}
}

func (m *TestMockPacketRepository) SavePacket(packet pkg.AppPacket) error {
	return nil
}

func (m *TestMockPacketRepository) SavePackets(packets []pkg.AppPacket) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Make a copy of the packets slice
	packetsCopy := make([]pkg.AppPacket, len(packets))
	copy(packetsCopy, packets)
	m.savedPackets = append(m.savedPackets, packetsCopy)

	if m.saveCalled != nil {
		select {
		case m.saveCalled <- struct{}{}:
		default:
		}
	}

	return m.savePacketsErr
}

func (m *TestMockPacketRepository) GetPackets(limit int, sort string) ([]pkg.SavedPacket, error) {
	return nil, nil
}

func (m *TestMockPacketRepository) GetPacket(packetID string) (pkg.SavedPacket, error) {
	return pkg.SavedPacket{}, nil
}

func (m *TestMockPacketRepository) DeletePacket(packetID string) error {
	return nil
}

func (m *TestMockPacketRepository) UpdatePacket(packet pkg.AppPacket) error {
	return nil
}

func (m *TestMockPacketRepository) GetSavedPackets() [][]pkg.AppPacket {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.savedPackets
}

func TestSniffAndStorePackets_BatchesSaveWhenCacheExceeds100(t *testing.T) {
	// Save original queue and restore after test
	originalQueue := pkg.PacketsToCaptureQueue
	defer func() { pkg.PacketsToCaptureQueue = originalQueue }()

	// Create a new queue for this test
	pkg.PacketsToCaptureQueue = pkg.PacketQueue{
		ItemsChan: make(chan pkg.AppPacket, 200),
	}

	// Create mock repository
	mockRepo := &TestMockPacketRepository{
		savedPackets: make([][]pkg.AppPacket, 0),
		saveCalled:   make(chan struct{}, 1),
	}

	// Start the sniff and store service
	SniffAndStorePackets(mockRepo)

	// Send more than 100 packets to trigger a save
	for i := 0; i < 101; i++ {
		pkg.PacketsToCaptureQueue.ItemsChan <- pkg.AppPacket{
			DeviceID:  "test-device",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// Wait for save to be called
	select {
	case <-mockRepo.saveCalled:
		// Save was called
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for SavePackets to be called")
	}

	// Check that packets were saved
	savedBatches := mockRepo.GetSavedPackets()
	if len(savedBatches) == 0 {
		t.Fatal("expected at least one batch of packets to be saved")
	}

	// The first batch should have 101 packets (cache > 100 triggers save)
	if len(savedBatches[0]) != 101 {
		t.Errorf("expected first batch to have 101 packets, got %d", len(savedBatches[0]))
	}
}

func TestSniffAndStorePackets_ProcessesMultipleBatches(t *testing.T) {
	// Save original queue and restore after test
	originalQueue := pkg.PacketsToCaptureQueue
	defer func() { pkg.PacketsToCaptureQueue = originalQueue }()

	// Create a new queue for this test
	pkg.PacketsToCaptureQueue = pkg.PacketQueue{
		ItemsChan: make(chan pkg.AppPacket, 300),
	}

	// Create mock repository
	mockRepo := &TestMockPacketRepository{
		savedPackets: make([][]pkg.AppPacket, 0),
		saveCalled:   make(chan struct{}, 10),
	}

	// Start the sniff and store service
	SniffAndStorePackets(mockRepo)

	// Send enough packets for two batches (101 + 101 = 202 packets)
	for i := 0; i < 202; i++ {
		pkg.PacketsToCaptureQueue.ItemsChan <- pkg.AppPacket{
			DeviceID:  "test-device",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// Wait for both saves to be called
	saveCount := 0
	timeout := time.After(3 * time.Second)
	for saveCount < 2 {
		select {
		case <-mockRepo.saveCalled:
			saveCount++
		case <-timeout:
			t.Fatalf("timeout waiting for SavePackets, only got %d calls", saveCount)
		}
	}

	// Check that two batches were saved
	savedBatches := mockRepo.GetSavedPackets()
	if len(savedBatches) < 2 {
		t.Errorf("expected at least 2 batches, got %d", len(savedBatches))
	}
}

func TestSniffAndStorePackets_DoesNotSaveUntilThreshold(t *testing.T) {
	// Save original queue and restore after test
	originalQueue := pkg.PacketsToCaptureQueue
	defer func() { pkg.PacketsToCaptureQueue = originalQueue }()

	// Create a new queue for this test
	pkg.PacketsToCaptureQueue = pkg.PacketQueue{
		ItemsChan: make(chan pkg.AppPacket, 200),
	}

	// Create mock repository
	mockRepo := &TestMockPacketRepository{
		savedPackets: make([][]pkg.AppPacket, 0),
		saveCalled:   make(chan struct{}, 1),
	}

	// Start the sniff and store service
	SniffAndStorePackets(mockRepo)

	// Send fewer than 101 packets
	for i := 0; i < 50; i++ {
		pkg.PacketsToCaptureQueue.ItemsChan <- pkg.AppPacket{
			DeviceID:  "test-device",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// Wait a bit and check that no save was called
	select {
	case <-mockRepo.saveCalled:
		t.Fatal("SavePackets should not be called when cache has fewer than 100 packets")
	case <-time.After(500 * time.Millisecond):
		// Expected - no save should happen
	}

	savedBatches := mockRepo.GetSavedPackets()
	if len(savedBatches) != 0 {
		t.Errorf("expected no batches saved, got %d", len(savedBatches))
	}
}

func TestSniffAndStorePackets_PreservesPacketData(t *testing.T) {
	// Save original queue and restore after test
	originalQueue := pkg.PacketsToCaptureQueue
	defer func() { pkg.PacketsToCaptureQueue = originalQueue }()

	// Create a new queue for this test
	pkg.PacketsToCaptureQueue = pkg.PacketQueue{
		ItemsChan: make(chan pkg.AppPacket, 200),
	}

	// Create mock repository
	mockRepo := &TestMockPacketRepository{
		savedPackets: make([][]pkg.AppPacket, 0),
		saveCalled:   make(chan struct{}, 1),
	}

	// Start the sniff and store service
	SniffAndStorePackets(mockRepo)

	// Send packets with specific device IDs
	expectedDeviceIDs := make([]string, 101)
	for i := 0; i < 101; i++ {
		deviceID := "device-" + string(rune('A'+i%26))
		expectedDeviceIDs[i] = deviceID
		pkg.PacketsToCaptureQueue.ItemsChan <- pkg.AppPacket{
			DeviceID:  deviceID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// Wait for save to be called
	select {
	case <-mockRepo.saveCalled:
		// Save was called
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for SavePackets to be called")
	}

	// Check that the packet data was preserved
	savedBatches := mockRepo.GetSavedPackets()
	if len(savedBatches) == 0 {
		t.Fatal("expected at least one batch")
	}

	for i, packet := range savedBatches[0] {
		if packet.DeviceID != expectedDeviceIDs[i] {
			t.Errorf("packet %d: expected device ID '%s', got '%s'", i, expectedDeviceIDs[i], packet.DeviceID)
		}
	}
}

func TestSniffAndStorePackets_StartsGoroutine(t *testing.T) {
	// Save original queue and restore after test
	originalQueue := pkg.PacketsToCaptureQueue
	defer func() { pkg.PacketsToCaptureQueue = originalQueue }()

	// Create a new queue for this test
	pkg.PacketsToCaptureQueue = pkg.PacketQueue{
		ItemsChan: make(chan pkg.AppPacket, 10),
	}

	mockRepo := &TestMockPacketRepository{
		savedPackets: make([][]pkg.AppPacket, 0),
	}

	// The function should return immediately (non-blocking)
	done := make(chan struct{})
	go func() {
		SniffAndStorePackets(mockRepo)
		close(done)
	}()

	select {
	case <-done:
		// Function returned immediately as expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("SniffAndStorePackets should return immediately (it starts a goroutine)")
	}
}
