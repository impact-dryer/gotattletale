package pkg

type PacketQueue struct {
	ItemsChan chan AppPacket
}

func (q *PacketQueue) Push(item AppPacket) {
	q.ItemsChan <- item
}

func (q *PacketQueue) PushMultiple(items []AppPacket) {
	for _, item := range items {
		q.ItemsChan <- item
	}
}
