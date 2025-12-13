package db

import "errors"

var ErrNotFound = errors.New("record not found")

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
