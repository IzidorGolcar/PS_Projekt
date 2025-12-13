package db

import (
	"errors"
	"seminarska/internal/data/storage/entities"
)

func (r *Relation[E]) Insert(e E) (Receipt, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.insertUnsafe(e)
}

func (r *Relation[E]) insertUnsafe(e E) (Receipt, error) {
	if e.Id() == 0 {
		e.SetId(r.idIndex.next())
	}
	if _, ok := r.records[e.Id()]; ok {
		return nil, errors.New("duplicate id")
	}
	err := r.uniqueIndex.Add(e)
	if err != nil {
		return nil, err
	}
	record := NewMutableRecord[E]()
	_ = record.Write(e)
	r.records[e.Id()] = record
	receipt := newInsertReceipt(r, record)
	return receipt, nil
}

type insertReceipt[E entities.Entity] struct {
	r      *Relation[E]
	record *MutableRecord[E]
}

func newInsertReceipt[E entities.Entity](r *Relation[E], record *MutableRecord[E]) *insertReceipt[E] {
	return &insertReceipt[E]{r: r, record: record}
}

func (i *insertReceipt[E]) Confirm() error {
	i.r.mx.Lock()
	defer i.r.mx.Unlock()
	if err := i.record.Commit(); err != nil {
		panic(err)
	}
	return nil
}

func (i *insertReceipt[E]) Cancel(err error) {
	i.r.mx.Lock()
	defer i.r.mx.Unlock()
	if err := i.record.Rollback(); err != nil {
		panic(err)
	}
	e, err := i.record.Value()
	if err != nil {
		panic(err)
	}
	delete(i.r.records, e.Id())
	i.r.uniqueIndex.Remove(e)
}
