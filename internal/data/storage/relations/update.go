package relations

import (
	"errors"
	"log"
	"seminarska/internal/data/storage/entities"
)

func (r *Relation[E]) Update(id int64, transform TransformFunc[E]) (Receipt, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	record, err := r.getRecord(id)
	if err != nil {
		return nil, err
	}

	updated := transform(record.value)
	if updated.Id() != id {
		return nil, errors.New("cannot change id")
	}
	receipt := newUpdateReceipt(r, record, updated)
	return receipt, nil
}

type updateReceipt[E entities.Entity] struct {
	r        *Relation[E]
	record   *record[E]
	newValue E
}

func (u *updateReceipt[E]) Confirm() error {
	u.r.mx.Lock()
	defer u.r.mx.Unlock()
	defer u.record.dirtyMx.Unlock()
	err := u.r.uniqueIndex.Replace(u.record.value, u.newValue)
	if err != nil {
		return err
	}
	u.record.value = u.newValue
	u.record.dirty = false
	return nil
}

func (u *updateReceipt[E]) Cancel(err error) {
	u.r.mx.Lock()
	u.record.dirty = false
	u.r.mx.Unlock()
	u.record.dirtyMx.Unlock()
	log.Println("Cancelled update:", err)
}

func newUpdateReceipt[E entities.Entity](
	r *Relation[E], record *record[E], newValue E,
) *updateReceipt[E] {
	record.dirty = true
	record.dirtyMx.Lock()
	return &updateReceipt[E]{r, record, newValue}
}
