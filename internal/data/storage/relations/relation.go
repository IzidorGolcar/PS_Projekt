package relations

import (
	"seminarska/internal/data/storage/entities"
	"seminarska/internal/data/storage/keys"
	"sync"
)

type Relation[E entities.Entity] struct {
	mx          *sync.RWMutex
	uniqueIndex *keys.Index
	records     map[int64]*record[E]
	idIndex     *monotonicIndex
}

type Result[E entities.Entity] struct {
	confirmed bool
	record    E
}

type TransformFunc[E entities.Entity] func(E) E

func NewRelation[E entities.Entity]() *Relation[E] {
	return &Relation[E]{
		mx:          &sync.RWMutex{},
		uniqueIndex: newUniqueIndex[E](),
		records:     make(map[int64]*record[E]),
		idIndex:     newMonotonicIndex(0),
	}
}
