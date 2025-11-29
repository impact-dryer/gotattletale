package pkg

import "testing"

func TestPacketQueue(t *testing.T) {
	queue := PacketQueue{}
	queue.Push(1)
	queue.Push(2)
	queue.Push(3)

	item, err := queue.Pop()
	if item != 1 {
		t.Errorf("Expected 1, got %d", item)
	}

	item, err = queue.Pop()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if item != 2 {
		t.Errorf("Expected 2, got %d", item)
	}

	item, err = queue.Pop()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if item != 3 {
		t.Errorf("Expected 3, got %d", item)
	}
}

func TestEmptyQueue(t *testing.T) {
	queue := PacketQueue{}

	_, err := queue.Pop()
	if err == nil {
		t.Errorf("Expected to get an error, got nil")
	}

}

func TestPopMultiple(t *testing.T) {
	queue := PacketQueue{}

	_, err := queue.PopMultiple(2)
	if err == nil {
		t.Errorf("Expected to get an error, got nil")
	}
}

func TestPopMultipleMultipleItems(t *testing.T) {
	queue := PacketQueue{}
	queue.Push(1)
	queue.PushMultiple([]interface{}{2, 3})

	items, err := queue.PopMultiple(2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
	if items[0] != 1 {
		t.Errorf("Expected 1, got %d", items[0])
	}
	if items[1] != 2 {
		t.Errorf("Expected 2, got %d", items[1])
	}
}
