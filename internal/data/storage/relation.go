package storage

import (
	"errors"
	"log"
	"sync"
)

type defaultRelation[T Record] struct {
	mx        *sync.RWMutex
	confirmed map[RecordId]T
	pending   map[uint64]T
}

func newDefaultRelation[T Record]() *defaultRelation[T] {
	return &defaultRelation[T]{
		mx:        &sync.RWMutex{},
		confirmed: make(map[RecordId]T),
		pending:   make(map[uint64]T),
	}
}

func (d *defaultRelation[T]) Insert(t T) (Receipt, error) {
	d.mx.RLock()
	defer d.mx.RUnlock()
	if _, ok := d.pending[t.hash()]; ok {
		return nil, errors.New("pending record already exists")
	}
	d.pending[t.hash()] = t
	receipt := newCallbackReceipt[T](t, d.confirm, d.cancel)
	return receipt, nil
}

func (d *defaultRelation[T]) confirm(record T, id RecordId) error {
	d.mx.Lock()
	defer d.mx.Unlock()
	delete(d.pending, record.hash())
	if _, ok := d.confirmed[id]; ok {
		return errors.New("cannot confirm record: duplicate id")
	}
	d.confirmed[record.Id()] = record
	return nil
}

func (d *defaultRelation[T]) cancel(record T, err error) {
	d.mx.Lock()
	defer d.mx.Unlock()
	log.Println("Canceling data write: ", err)
	delete(d.pending, record.hash())
}

func (d *defaultRelation[T]) Get(id RecordId) (T, error) {
	if record, ok := d.confirmed[id]; !ok {
		var r T
		return r, errors.New("cannot get record: not found")
	} else {
		return record, nil
	}
}
