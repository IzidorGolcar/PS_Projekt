package relations

import (
	"seminarska/internal/data/storage/entities"
	"sync"
)

type record[E entities.Entity] struct {
	value   E
	dirtyMx *sync.Mutex
	deleted bool
	dirty   bool
}

func newRecord[E entities.Entity](value E) *record[E] {
	return &record[E]{
		value:   value,
		dirtyMx: &sync.Mutex{},
	}
}
