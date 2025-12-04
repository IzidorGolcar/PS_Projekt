package storage

import "errors"

type keyIndex map[uint64]any

func (i keyIndex) add(t Record) error {
	key := primaryKeyHash(t)
	if _, used := i[key]; used {
		return errors.New("unique constraint violation")
	}
	i[key] = t
	return nil
}

func (i keyIndex) remove(t Record) {
	delete(i, primaryKeyHash(t))
}
