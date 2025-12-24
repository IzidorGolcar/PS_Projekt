package db

import (
	"errors"
	"seminarska/internal/data/storage/entities"
)

func (r *Relation[E]) Insert(e E) (*InsertReceipt[E], error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	receipt, err := r.insertUnsafe(e)
	return NewInsertReceipt(receipt), err
}

func (r *Relation[E]) insertUnsafe(e E) (*UnsafeInsertReceipt[E], error) {
	err := r.uniqueIndex.Add(e)
	if err != nil {
		return nil, err
	}
	record := NewMutableRecord[E]()
	_ = record.Write(e)
	r.records[e.Id()] = record
	receipt := newUnsafeInsertReceipt(r, record)
	return receipt, nil
}

type UnsafeInsertReceipt[E entities.Entity] struct {
	r      *Relation[E]
	record *MutableRecord[E]
}

func newUnsafeInsertReceipt[E entities.Entity](r *Relation[E], record *MutableRecord[E]) *UnsafeInsertReceipt[E] {
	return &UnsafeInsertReceipt[E]{r: r, record: record}
}

func (i *UnsafeInsertReceipt[E]) Confirm() error {
	if err := i.record.Commit(); err != nil {
		panic(err)
	}
	return nil
}

func (i *UnsafeInsertReceipt[E]) Cancel(err error) {
	if err := i.record.Rollback(); err != nil {
		panic(err)
	}
	e, err := i.record.Value()
	if err != nil && !errors.Is(err, ErrUninitialized) {
		panic(err)
	}
	delete(i.r.records, e.Id())
	i.r.uniqueIndex.Remove(e)
}

type InsertReceipt[E entities.Entity] struct {
	*UnsafeInsertReceipt[E]
}

func NewInsertReceipt[E entities.Entity](unsafeInsertReceipt *UnsafeInsertReceipt[E]) *InsertReceipt[E] {
	return &InsertReceipt[E]{UnsafeInsertReceipt: unsafeInsertReceipt}
}

func (i *InsertReceipt[E]) Confirm() error {
	i.r.mx.Lock()
	defer i.r.mx.Unlock()
	return i.UnsafeInsertReceipt.Confirm()
}

func (i *InsertReceipt[E]) Cancel(err error) {
	i.r.mx.Lock()
	defer i.r.mx.Unlock()
	i.UnsafeInsertReceipt.Cancel(err)
}
