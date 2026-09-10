package broadcaster

import (
	"fmt"
	"sync"
)

type Broadcaster struct {
	mu      sync.RWMutex
	clients map[chan string]bool
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[chan string]bool),
	}
}

func (b *Broadcaster) Subscribe() chan string {
	ch := make(chan string, 16)
	b.mu.Lock()
	b.clients[ch] = true
	b.mu.Unlock()
	return ch
}

func (b *Broadcaster) Unsubscribe(ch chan string) {
	b.mu.Lock()
	if _, ok := b.clients[ch]; ok {
		delete(b.clients, ch)
		close(ch)
	}
	b.mu.Unlock()
}

func (b *Broadcaster) Publish(event string, data string) {
	payload := fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.clients {
		select {
		case ch <- payload:
		default:
			// Non-blocking write: if client buffer is full, skip to avoid slow consumer bottleneck
		}
	}
}
