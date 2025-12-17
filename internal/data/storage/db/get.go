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

func (r *Relation[E]) Get(id int64) (e E, err error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	record, err := r.getRecord(id)
	if err != nil {
		return
	}
	if record.IsDirty() {
		err = ErrNotFound
		return
	}
	return record.Value()
}

const NoLimit = 0

func (r *Relation[E]) GetPredicate(
	predicate PredicateFunc[E],
	limit int,
) (values []E, err error) {
	if limit < 0 {
		return nil, errors.New("limit must be non-negative")
	}
	r.mx.RLock()
	defer r.mx.RUnlock()
	for _, record := range r.records {
		e, err := record.Value()
		if err != nil {
			continue
		}
		if predicate(e) {
			values = append(values, e)
		}
		if len(values) == limit && limit != NoLimit {
			break
		}
	}
	return
}

func (r *Relation[E]) GetAll() (values []E, err error) {
	return r.GetPredicate(func(e E) bool { return true }, NoLimit)
}

func (r *Relation[E]) Count() int {
	r.mx.RLock()
	defer r.mx.RUnlock()
	return len(r.records)
}

func (r *Relation[E]) GetTransform(id int64, transform TransformFunc[E]) (e E, err error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	return r.getTransform(id, transform)
}
