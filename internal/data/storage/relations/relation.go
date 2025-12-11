package relations

import (
	"errors"
	"log"
	"seminarska/internal/data/storage/entities"
	"seminarska/internal/data/storage/keys"
	"sync"
)

type Relation[E entities.Entity] struct {
	mx          *sync.RWMutex
	uniqueIndex *keys.Index
	confirmed   map[int64]E
	pending     map[int64]E
}

type Result[T entities.Entity] struct {
	confirmed bool
	record    T
}

func NewRelation[E entities.Entity]() *Relation[E] {
	return &Relation[E]{
		mx:          &sync.RWMutex{},
		uniqueIndex: newUniqueIndex[E](),
		confirmed:   make(map[int64]E),
		pending:     make(map[int64]E),
	}
}

func (r *Relation[E]) Insert(e E) (Receipt, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	if e.Id() == 0 {
		e.SetId(r.newAutoId())
	}
	err := r.uniqueIndex.Add(e)
	if err != nil {
		return nil, err
	}
	if _, ok := r.pending[e.Id()]; ok {
		return nil, errors.New("duplicate id")
	}
	if _, ok := r.confirmed[e.Id()]; ok {
		return nil, errors.New("duplicate id")
	}
	r.pending[e.Id()] = e
	receipt := newCallbackReceipt[E](e, r.confirm, r.cancel)
	return receipt, nil
}

func (r *Relation[E]) newAutoId() int64 {
	return int64(len(r.confirmed)) + int64(len(r.pending))
}

func (r *Relation[E]) confirm(e E) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	delete(r.pending, e.Id())
	if _, ok := r.confirmed[e.Id()]; ok {
		panic("illegal state")
	}
	r.confirmed[e.Id()] = e
	return nil
}

func (r *Relation[E]) cancel(e E, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	log.Println("Canceling data write:", err)
	// TODO
	panic("not implemented")
}

func (r *Relation[E]) getUnsafe(id int64) (res Result[E], err error) {
	var ok bool
	if res.record, ok = r.confirmed[id]; ok {
		res.confirmed = true
		return
	} else if res.record, ok = r.pending[id]; ok {
		res.confirmed = false
		return
	}
	err = errors.New("not found")
	return
}

func (r *Relation[E]) Get(id int64) (res Result[E], err error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	return r.getUnsafe(id)
}

func (r *Relation[E]) Delete(id int64) error {
	// fixme can delete pending message (must integrate with chain, to check if tail has the record)
	r.mx.Lock()
	defer r.mx.Unlock()
	e, err := r.getUnsafe(id)
	if err != nil {
		return err
	}
	if e.confirmed {
		r.uniqueIndex.Remove(e.record)
		delete(r.confirmed, id)
	} else {
		return errors.New("cannot delete pending record")
	}
	return nil
}
