package pkg

import "testing"

func TestPacketQueuePush(t *testing.T) {
	q := PacketQueue{
		ItemsChan: make(chan AppPacket, 1),
	}

	want := AppPacket{DeviceID: "eth0"}
	q.Push(want)

	select {
	case got := <-q.ItemsChan:
		if got.DeviceID != want.DeviceID {
			t.Fatalf("expected device %s, got %s", want.DeviceID, got.DeviceID)
		}
	default:
		t.Fatal("expected packet in queue, found none")
	}
}

func TestPacketQueuePushMultiple(t *testing.T) {
	q := PacketQueue{
		ItemsChan: make(chan AppPacket, 2),
	}
	packets := []AppPacket{
		{DeviceID: "eth0"},
		{DeviceID: "eth1"},
	}

	q.PushMultiple(packets)

	gotFirst := <-q.ItemsChan
	if gotFirst.DeviceID != "eth0" {
		t.Fatalf("unexpected first packet: %s", gotFirst.DeviceID)
	}
	gotSecond := <-q.ItemsChan
	if gotSecond.DeviceID != "eth1" {
		t.Fatalf("unexpected second packet: %s", gotSecond.DeviceID)
	}
}

