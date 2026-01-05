package chain

import (
	"errors"
	"math"
	"strconv"
	"sync"
)

type Indexable interface {
	GetMessageIndex() int32
}

const MaxSize = math.MaxInt

var (
	ErrIncompleteResult   = errors.New("incomplete result")
	ErrIndexOutOfOrder    = errors.New("index out of order")
	ErrNoBufferedMessages = errors.New("no buffered messages")
)

type ReplayBuffer[T Indexable] struct {
	buffer []T
	mx     *sync.RWMutex
	size   int
}

func NewReplayBuffer[T Indexable](size int) *ReplayBuffer[T] {
	return &ReplayBuffer[T]{mx: &sync.RWMutex{}, size: size}
}

func (b *ReplayBuffer[T]) Add(messages ...T) error {
	b.mx.Lock()
	defer b.mx.Unlock()
	for _, msg := range messages {
		if len(b.buffer) > 0 && msg.GetMessageIndex() <= b.buffer[len(b.buffer)-1].GetMessageIndex() {
			return errors.Join(ErrIndexOutOfOrder, errors.New(strconv.Itoa(int(msg.GetMessageIndex()))))
		}
		b.buffer = append(b.buffer, msg)
		if len(b.buffer) > b.size {
			b.buffer = b.buffer[1:]
		}
	}
	return nil
}

func (b *ReplayBuffer[T]) LastMessageIndex() (int32, error) {
	b.mx.RLock()
	defer b.mx.RUnlock()
	if len(b.buffer) == 0 {
		return 0, ErrNoBufferedMessages
	}
	return b.buffer[len(b.buffer)-1].GetMessageIndex(), nil
}

func (b *ReplayBuffer[T]) MessagesAfter(index int32) ([]T, error) {
	b.mx.RLock()
	defer b.mx.RUnlock()
	for i, msg := range b.buffer {
		if msg.GetMessageIndex() == index {
			if i == len(b.buffer)-1 {
				return []T{}, nil
			}
			return b.buffer[i+1:], nil
		}
		if msg.GetMessageIndex() == index+1 {
			return b.buffer[i:], nil
		}
		if msg.GetMessageIndex() > index {
			return b.buffer[i:], ErrIncompleteResult
		}
	}
	return nil, ErrNoBufferedMessages
}

func (b *ReplayBuffer[T]) ClearBefore(index int32) {
	b.mx.Lock()
	defer b.mx.Unlock()
	for i, msg := range b.buffer {
		if msg.GetMessageIndex() >= index {
			b.buffer = b.buffer[i:]
			return
		}
	}
}
