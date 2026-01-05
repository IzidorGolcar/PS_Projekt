package broadcast

import (
	"context"
	"sync"
)

type Broadcaster[T any] struct {
	mu   sync.RWMutex
	subs map[chan T]struct{}
}

func New[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		subs: make(map[chan T]struct{}),
	}
}

func (b *Broadcaster[T]) Subscribe(ctx context.Context) <-chan T {
	ch := make(chan T, 1)

	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()

		b.mu.Lock()
		delete(b.subs, ch)
		b.mu.Unlock()

		close(ch)
	}()

	return ch
}

// Broadcast sends v to all active subscribers.
func (b *Broadcaster[T]) Broadcast(v T) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subs {
		select {
		case ch <- v:
		default:
			// drop or handle slow subscriber
		}
	}
}
