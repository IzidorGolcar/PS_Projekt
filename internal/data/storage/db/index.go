package db

import "sync"

type monotonicIndex struct {
	value int64
	mx    *sync.Mutex
}

func (i *monotonicIndex) next() int64 {
	i.mx.Lock()
	defer i.mx.Unlock()
	current := i.value
	i.value++
	return current
}

func newMonotonicIndex(initial int64) *monotonicIndex {
	return &monotonicIndex{
		value: initial,
		mx:    &sync.Mutex{},
	}
}
