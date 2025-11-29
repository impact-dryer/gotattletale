package pkg

import (
	"errors"
	"sync"
)

type PacketQueue struct {
	sync.Mutex
	Items []AppPacket
}

func (q *PacketQueue) Push(item AppPacket) {
	q.Lock()
	defer q.Unlock()
	q.Items = append(q.Items, item)
}

func (q *PacketQueue) Pop() (AppPacket, error) {
	q.Lock()
	defer q.Unlock()
	if len(q.Items) == 0 {
		return AppPacket{}, errors.New("queue is empty")
	}
	packetToReturn := q.Items[0]
	q.Items = q.Items[1:]
	return packetToReturn, nil
}

func (q *PacketQueue) PushMultiple(items []AppPacket) {
	q.Lock()
	defer q.Unlock()
	q.Items = append(q.Items, items...)
}

func (q *PacketQueue) PopMultiple(count int) ([]AppPacket, error) {
	q.Lock()
	defer q.Unlock()
	if len(q.Items) == 0 {
		return nil, errors.New("queue is empty")
	}
	if count > len(q.Items) {
		return nil, errors.New("count is greater than the number of items in the queue")
	}
	packetsToReturn := q.Items[:count]
	q.Items = q.Items[count:]
	return packetsToReturn, nil
}
