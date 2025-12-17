package db

import (
	"seminarska/internal/data/storage/entities"
)

func (r *Relation[E]) Insert(e E) (*InsertReceipt[E], error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.insertUnsafe(e)
}

func (r *Relation[E]) insertUnsafe(e E) (*InsertReceipt[E], error) {
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

type InsertReceipt[E entities.Entity] struct {
	r      *Relation[E]
	record *MutableRecord[E]
}

func newInsertReceipt[E entities.Entity](r *Relation[E], record *MutableRecord[E]) *InsertReceipt[E] {
	return &InsertReceipt[E]{r: r, record: record}
}

func (i *InsertReceipt[E]) Confirm() error {
	i.r.mx.Lock()
	defer i.r.mx.Unlock()
	if err := i.record.Commit(); err != nil {
		panic(err)
	}
	return nil
}

func (i *InsertReceipt[E]) Cancel(err error) {
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
