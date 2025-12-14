package db

import (
	"errors"
	"seminarska/internal/data/storage/entities"
)

var ErrNotFound = errors.New("record not found")

type PredicateFunc[E entities.Entity] func(E) bool

func (r *Relation[E]) getRecord(id int64) (e *MutableRecord[E], err error) {
	rec, ok := r.records[id]
	if !ok {
		err = ErrNotFound
		return
	}
	return rec, nil
}

func (r *Relation[E]) Get(id int64) (rec Record[E], err error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	record, err := r.getRecord(id)
	if err != nil {
		return
	}
	return record.CurrentSnapshot(), nil
}

func (r *Relation[E]) GetPredicate(
	predicate PredicateFunc[E],
	limit int,
) (rec []Record[E], err error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	for _, record := range r.records {
		if record.IsDirty() {
			e, err := record.DirtyValue()
			if err != nil {
				continue
			}
			if predicate(e) {
				rec = append(rec, record.CurrentSnapshot())
			}
		} else {
			e, err := record.Value()
			if err != nil {
				continue
			}
			if predicate(e) {
				rec = append(rec, record.CurrentSnapshot())
			}
		}
		if len(rec) == limit {
			break
		}
	}
	return
}
