package keys

import (
	"errors"
	"seminarska/internal/data/storage/entities"
)

var (
	ErrDuplicateId = errors.New("duplicate id")
	ErrConstraint  = errors.New("unique constraint violation")
)

type Index struct {
	set           map[uint64]struct{}
	ids           map[int64]struct{}
	indexedFields []string
}

func NewIndex(fields ...string) *Index {
	return &Index{
		set:           make(map[uint64]struct{}),
		ids:           make(map[int64]struct{}),
		indexedFields: fields,
	}
}

func (i Index) Add(e entities.Entity) error {
	if _, used := i.ids[e.Id()]; used {
		return ErrDuplicateId
	}
	if len(i.indexedFields) == 0 {
		return nil
	}
	key := structHash(e, i.indexedFields)
	if _, used := i.set[key]; used {
		return ErrConstraint
	}
	i.set[key] = struct{}{}
	return nil
}

func (i Index) Remove(e entities.Entity) {
	delete(i.ids, e.Id())
	if len(i.indexedFields) > 0 {
		delete(i.set, structHash(e, i.indexedFields))
	}
}

func (i Index) Replace(old, new entities.Entity) error {
	if old.Id() != new.Id() {
		return errors.New("cannot replace entities with different ids")
	}
	current := structHash(old, i.indexedFields)
	i.Remove(old)
	err := i.Add(new)
	if err != nil {
		i.set[current] = struct{}{}
	}
	return err
}

func (i Index) Reset() {
	i.set = make(map[uint64]struct{})
	i.ids = make(map[int64]struct{})
}
