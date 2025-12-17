package db

import (
	"errors"
	"seminarska/internal/data/storage/entities"
)

type TransformFunc[E entities.Entity] func(E) (E, error)

var ErrTransform = errors.New("failed to transform")
var ErrIdChanged = errors.New("id changed")

func (r *Relation[E]) getTransform(id int64, transform TransformFunc[E]) (e E, err error) {
	record, err := r.getRecord(id)
	if err != nil {
		return
	}
	current, err := record.Value()
	if err != nil {
		return
	}
	updated, err := transform(current)
	if err != nil {
		err = errors.Join(ErrTransform, err)
		return
	}
	if updated.Id() != id {
		err = errors.Join(ErrIdChanged, err)
		return
	}
	return updated, nil
}
