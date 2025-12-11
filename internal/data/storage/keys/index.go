package keys

import (
	"errors"
)

type Index struct {
	set           map[uint64]struct{}
	indexedFields []string
}

func NewIndex(fields ...string) *Index {
	return &Index{
		set:           make(map[uint64]struct{}),
		indexedFields: fields,
	}
}

func (i Index) Add(t any) error {
	if len(i.indexedFields) == 0 {
		return nil
	}
	key := structHash(t, nil)
	if _, used := i.set[key]; used {
		return errors.New("unique constraint violation")
	}
	i.set[key] = struct{}{}
	return nil
}

func (i Index) Remove(t any) {
	if len(i.indexedFields) > 0 {
		delete(i.set, structHash(t, i.indexedFields))
	}
}
