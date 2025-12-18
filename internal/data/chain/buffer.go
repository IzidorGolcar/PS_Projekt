package chain

import (
	"context"
	"errors"
	"slices"
	"sync"
)

type Indexable interface {
	GetMessageIndex() int32
}

type ReplayBuffer[T Indexable] struct {
	ctx    context.Context
	src    <-chan T
	buffer []T
	dst    chan T
	mx     *sync.RWMutex
}

func NewBuffer[T Indexable](ctx context.Context, src <-chan T) *ReplayBuffer[T] {
	b := &ReplayBuffer[T]{
		ctx: ctx,
		src: src,
		dst: make(chan T, 1000),
		mx:  &sync.RWMutex{},
	}
	go b.run()
	return b
}

func (b *ReplayBuffer[T]) run() {
	for {
		select {
		case <-b.ctx.Done():
			return
		case v := <-b.src:
			b.dst <- v
			b.mx.Lock()
			b.buffer = append(b.buffer, v)
			b.mx.Unlock()
		}
	}
}

func (b *ReplayBuffer[T]) Messages() <-chan T {
	return b.dst
}

func (b *ReplayBuffer[T]) Clear() {
	b.mx.Lock()
	defer b.mx.Unlock()
	b.buffer = nil
}

func (b *ReplayBuffer[T]) LastMessageIndex() (i int32, err error) {
	b.mx.RLock()
	defer b.mx.RUnlock()
	if len(b.buffer) == 0 {
		err = errors.New("no buffered messages")
		return
	}
	i = b.buffer[len(b.buffer)-1].GetMessageIndex()
	return
}

func (b *ReplayBuffer[T]) MessagesAfter(index int32) (out []T) {
	for _, msg := range b.buffer {
		if msg.GetMessageIndex() > index {
			out = append(out, msg)
		}
	}
	slices.SortFunc(out, func(a, b T) int {
		return int(a.GetMessageIndex()) - int(b.GetMessageIndex())
	})
	return
}
