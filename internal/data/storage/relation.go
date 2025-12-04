package storage

import (
	"errors"
	"log"
	"sync"
)

type relation[T Record] struct {
	mx        *sync.RWMutex
	keys      keyIndex
	confirmed map[uint64]*T
	pending   map[uint64]*T
}

func newRelation[T Record]() *relation[T] {
	return &relation[T]{
		mx:        &sync.RWMutex{},
		keys:      keyIndex{},
		confirmed: make(map[uint64]*T),
		pending:   make(map[uint64]*T),
	}
}

func (r *relation[T]) Insert(t T) (Receipt, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	t.SetId(r.newAutoId())
	err := r.keys.add(t)
	if err != nil {
		return nil, err
	}
	r.pending[t.Id()] = &t
	receipt := newCallbackReceipt[T](t, r.confirm, r.cancel)
	return receipt, nil
}

func (r *relation[T]) newAutoId() uint64 {
	return uint64(len(r.confirmed)) + uint64(len(r.pending))
}

func (r *relation[T]) confirm(record T) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	delete(r.pending, record.Id())
	if _, ok := r.confirmed[record.Id()]; ok {
		panic("illegal state")
	}
	r.confirmed[record.Id()] = &record
	return nil
}

func (r *relation[T]) cancel(record T, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	r.keys.remove(record)
	log.Println("Canceling data write: ", err)
	delete(r.pending, record.Id())
}

func (r *relation[T]) Get(id uint64) (T, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	if record, ok := r.confirmed[id]; !ok {
		var r T
		return r, errors.New("no such record")
	} else {
		return *record, nil
	}
}
