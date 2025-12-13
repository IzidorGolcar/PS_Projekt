package db

import (
	"errors"
	"seminarska/internal/data/storage/entities"
)

func (r *Relation[E]) Delete(id int64) (*DeleteReceipt[E], error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	e, err := r.getRecord(id)
	if err != nil {
		return nil, err
	}
	e.Delete()
	receipt := newDeleteReceipt(r, e)
	return receipt, nil
}

type DeleteReceipt[E entities.Entity] struct {
	r      *Relation[E]
	record *MutableRecord[E]
}

func (d *DeleteReceipt[E]) DeletedValue() E {
	deleted, err := d.record.DirtyValue()
	if !errors.Is(err, ErrDeleted) {
		panic(err)
	}
	return deleted
}

func newDeleteReceipt[E entities.Entity](
	r *Relation[E],
	record *MutableRecord[E],
) *DeleteReceipt[E] {
	return &DeleteReceipt[E]{r, record}
}

func (d *DeleteReceipt[E]) Confirm() (err error) {
	d.r.mx.Lock()
	defer d.r.mx.Unlock()

	indexedValue, err := d.record.Value()
	if err != nil {
		panic(err)
	}

	if err := d.record.Commit(); err != nil {
		panic(err)
	}
	d.r.uniqueIndex.Remove(indexedValue)
	delete(d.r.records, indexedValue.Id())
	return nil
}

func (d *DeleteReceipt[E]) Cancel(err error) {
	d.r.mx.Lock()
	defer d.r.mx.Unlock()
	if err := d.record.Rollback(); err != nil {
		panic(err)
	}
}
