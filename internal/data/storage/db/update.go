package db

import (
	"errors"
	"seminarska/internal/data/storage/entities"
)

type TransformFunc[E entities.Entity] func(E) E

func (r *Relation[E]) Update(id int64, transform TransformFunc[E]) (*UpdateReceipt[E], error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	record, err := r.getRecord(id)
	if err != nil {
		return nil, err
	}
	current, err := record.Value()
	if err != nil {
		return nil, err
	}
	updated := transform(current)
	if updated.Id() != id {
		return nil, errors.New("cannot change id")
	}
	err = record.Write(updated)
	if err != nil {
		return nil, err
	}
	receipt := newUpdateReceipt(r, record)
	return receipt, nil
}

type UpdateReceipt[E entities.Entity] struct {
	r      *Relation[E]
	record *MutableRecord[E]
}

func newUpdateReceipt[E entities.Entity](
	r *Relation[E], record *MutableRecord[E],
) *UpdateReceipt[E] {
	return &UpdateReceipt[E]{r, record}
}

func (u *UpdateReceipt[E]) NewValue() E {
	newValue, err := u.record.DirtyValue()
	if err != nil {
		panic(err)
	}
	return newValue
}

func (u *UpdateReceipt[E]) Confirm() error {
	u.r.mx.Lock()
	defer u.r.mx.Unlock()
	err := u.r.uniqueIndex.Replace(
		u.record.confirmedValue.e,
		u.record.dirtyValue.e,
	)
	if err != nil {
		if err := u.record.Rollback(); err != nil {
			panic(err)
		}
		return err
	}
	if err := u.record.Commit(); err != nil {
		panic(err)
	}
	return nil
}

func (u *UpdateReceipt[E]) Cancel(err error) {
	u.r.mx.Lock()
	defer u.r.mx.Unlock()
	if err = u.record.Rollback(); err != nil {
		panic(err)
	}
}
