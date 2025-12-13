package relations

import "errors"

var ErrNotFound = errors.New("record not found")

func (r *Relation[E]) getRecord(id int64) (e *record[E], err error) {
	rec, ok := r.records[id]
	if !ok {
		err = ErrNotFound
		return
	}
	return rec, nil
}

// Get returns the current value and whether the record is currently dirty.
// It never blocks on per-record mutexes; it only takes the relation read lock.
func (r *Relation[E]) Get(id int64) (val E, dirty bool, err error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	rec, err := r.getRecord(id)
	if err != nil {
		return
	}
	if rec.deleted {
		err = ErrNotFound
		return
	}
	return rec.value, rec.dirty, nil
}
