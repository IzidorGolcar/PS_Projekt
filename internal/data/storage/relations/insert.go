package relations

import (
	"errors"
	"log"
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
	record := newRecord(e)
	r.records[e.Id()] = record
	receipt := newInsertReceipt(r, record)
	return receipt, nil
}

type insertReceipt[E entities.Entity] struct {
	r      *Relation[E]
	record *record[E]
}

func newInsertReceipt[E entities.Entity](r *Relation[E], record *record[E]) *insertReceipt[E] {
	record.dirty = true
	record.dirtyMx.Lock()
	return &insertReceipt[E]{r: r, record: record}
}

func (i *insertReceipt[E]) Confirm() error {
	i.r.mx.Lock()
	i.record.dirty = false
	i.r.mx.Unlock()
	i.record.dirtyMx.Unlock()
	return nil
}

func (i *insertReceipt[E]) Cancel(err error) {
	log.Println("Cancelled insert:", err)
	i.r.mx.Lock()
	defer i.r.mx.Unlock()
	i.record.dirty = false
	delete(i.r.records, i.record.value.Id())
	i.r.uniqueIndex.Remove(i.record.value)
	i.record.dirtyMx.Unlock()
}
