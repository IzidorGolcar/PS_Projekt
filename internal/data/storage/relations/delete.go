package relations

import (
	"log"
	"seminarska/internal/data/storage/entities"
)

func (r *Relation[E]) Delete(id int64) (Receipt, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	e, err := r.getRecord(id)
	if err != nil {
		return nil, err
	}
	receipt := newDeleteReceipt(r, e)
	return receipt, nil
}

type deleteReceipt[E entities.Entity] struct {
	r      *Relation[E]
	record *record[E]
}

func newDeleteReceipt[E entities.Entity](r *Relation[E], record *record[E]) *deleteReceipt[E] {
	// Relation mutex is held by caller when creating the receipt
	record.dirty = true
	record.dirtyMx.Lock()
	return &deleteReceipt[E]{r: r, record: record}
}

func (d *deleteReceipt[E]) Confirm() error {
	d.r.mx.Lock()
	defer d.r.mx.Unlock()
	defer d.record.dirtyMx.Unlock()
	d.record.deleted = true
	d.r.uniqueIndex.Remove(d.record.value)
	delete(d.r.records, d.record.value.Id())
	return nil
}

func (d *deleteReceipt[E]) Cancel(err error) {
	d.r.mx.Lock()
	defer d.r.mx.Unlock()
	log.Println("Cancelled deletion:", err)
	d.record.dirty = false
	d.record.dirtyMx.Unlock()
}
