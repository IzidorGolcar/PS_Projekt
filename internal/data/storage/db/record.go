package db

import (
	"errors"
	"seminarska/internal/data/storage/entities"
	"sync"
)

var (
	ErrUninitialized = errors.New("confirmedValue not initialized")
	ErrDeleted       = errors.New("record deleted")
	ErrNotDirty      = errors.New("record not dirty")
)

type value[E entities.Entity] struct {
	e       E
	deleted bool
}

type Record[E entities.Entity] interface {
	IsDirty() bool
	Value() (e E, err error)
	DirtyValue() (e E, err error)
}

type SnapshotRecord[E entities.Entity] struct {
	initialized    bool
	dirty          bool
	confirmedValue value[E]
	dirtyValue     value[E]
}

func (s *SnapshotRecord[E]) IsDirty() bool {
	return s.dirty
}

func (s *SnapshotRecord[E]) Value() (e E, err error) {
	if !s.initialized {
		err = ErrUninitialized
	} else if s.confirmedValue.deleted {
		err = ErrDeleted
	} else {
		e = s.confirmedValue.e
	}
	return
}

func (s *SnapshotRecord[E]) DirtyValue() (e E, err error) {
	if !s.dirty {
		err = ErrNotDirty
	} else if s.dirtyValue.deleted {
		err = ErrDeleted
	} else {
		e = s.dirtyValue.e
	}
	return
}

func (s *SnapshotRecord[E]) Copy() *SnapshotRecord[E] {
	return &SnapshotRecord[E]{
		s.initialized,
		s.dirty,
		s.confirmedValue,
		s.dirtyValue,
	}
}

type MutableRecord[E entities.Entity] struct {
	dirtyMx *sync.Mutex
	SnapshotRecord[E]
}

func NewMutableRecord[E entities.Entity]() *MutableRecord[E] {
	return &MutableRecord[E]{
		dirtyMx: &sync.Mutex{},
	}
}

func (r *MutableRecord[E]) Delete() {
	r.dirtyMx.Lock()
	r.dirtyValue.e = r.confirmedValue.e
	r.dirtyValue.deleted = true
	r.dirty = true
}

func (r *MutableRecord[E]) Write(value E) error {
	r.dirtyMx.Lock()
	if r.confirmedValue.deleted {
		r.dirtyMx.Unlock()
		return ErrDeleted
	}
	r.dirty = true
	r.dirtyValue.e = value
	return nil
}

func (r *MutableRecord[E]) Commit() error {
	if !r.dirty {
		return ErrNotDirty
	}
	r.confirmedValue = r.dirtyValue
	r.initialized = true
	r.dirty = false
	r.dirtyMx.Unlock()
	return nil
}

func (r *MutableRecord[E]) Rollback() error {
	if !r.dirty {
		return ErrNotDirty
	}
	r.dirty = false
	r.dirtyValue = r.confirmedValue
	r.dirtyMx.Unlock()
	return nil
}

func (r *MutableRecord[E]) CurrentSnapshot() *SnapshotRecord[E] {
	return r.SnapshotRecord.Copy()
}
