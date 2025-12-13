package storage

import (
	"fmt"
	"seminarska/internal/data/storage/db"
	"seminarska/internal/data/storage/entities"
)

type ChainHandler interface {
	Forward(entities.Entity) error
	Delete(entities.Entity) error
	Update(entities.Entity) error
	Compare(entities.Entity) (bool, error)
}

func ChainedInsert[E entities.Entity](r *db.Relation[E], e E, chain ChainHandler) error {
	receipt, err := r.Insert(e)
	if err != nil {
		return fmt.Errorf("failed to insert e: %w", err)
	}
	err = chain.Forward(e)
	if err != nil {
		receipt.Cancel(err)
		return fmt.Errorf("failed to forward request: %w", err)
	}
	err = receipt.Confirm()
	if err != nil {
		return fmt.Errorf("failed to confirm receipt: %w", err)
	}
	return nil
}

func ChainedDelete[E entities.Entity](r *db.Relation[E], id int64, chain ChainHandler) error {
	rec, err := r.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	err = chain.Delete(rec.DeletedValue())
	if err != nil {
		rec.Cancel(err)
		return fmt.Errorf("failed to forward request: %w", err)
	}
	err = rec.Confirm()
	if err != nil {
		return fmt.Errorf("failed to confirm receipt: %w", err)
	}
	return nil

}

func ChainedUpdate[E entities.Entity](
	r *db.Relation[E],
	id int64,
	transform db.TransformFunc[E],
	chain ChainHandler,
) error {
	rec, err := r.Update(id, transform)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	err = chain.Update(rec.NewValue())
	if err != nil {
		return fmt.Errorf("failed to forward request: %w", err)
	}
	err = rec.Confirm()
	if err != nil {
		return fmt.Errorf("failed to confirm receipt: %w", err)
	}
	return nil
}

func ChainedGet[E entities.Entity](r *db.Relation[E], id int64, chain ChainHandler) (e E, err error) {
	record, err := r.Get(id)
	if err != nil {
		err = fmt.Errorf("failed to get record: %w", err)
		return
	}
	if record.IsDirty() {
		e, _ = record.DirtyValue()
		var eq bool
		eq, err = chain.Compare(e)
		if err != nil {
			err = fmt.Errorf("failed to sync record: %w", err)
		}
		if !eq {
			e, _ = record.Value()
		}
	}
	return
}
