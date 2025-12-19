package db

import "errors"

func (r *Relation[E]) Import(snapshot []E) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if len(r.records) != 0 {
		return errors.New("cannot import into non-empty relation")
	}

	r.uniqueIndex.Reset()
	for _, e := range snapshot {
		receipt, err := r.insertUnsafe(e)
		if err != nil {
			return err
		}
		err = receipt.Confirm()
		if err != nil {
			return err
		}
	}
	return nil
}
