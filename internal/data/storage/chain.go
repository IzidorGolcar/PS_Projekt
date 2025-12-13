package storage

import (
	"errors"
	"fmt"
	"seminarska/internal/data/storage/entities"
	"seminarska/internal/data/storage/relations"
	"seminarska/proto/datalink"
)

type DatabaseChain interface {
	ForwardRequest(*datalink.Record) error
	IsSynced(*datalink.Record) (bool, error)
}

func ChainedInsert[T entities.Entity](r *relations.Relation[T], record T, chain DatabaseChain) error {
	receipt, err := r.Insert(record)
	if err != nil {
		return fmt.Errorf("failed to insert record: %w", err)
	}
	err = chain.ForwardRequest(record.ToDatalinkRecord())
	if err != nil {
		receipt.Cancel(err)
		return fmt.Errorf("failed to forward request: %w", err)
	}
	err = receipt.Confirm()
	if err != nil {
		receipt.Cancel(err)
		return fmt.Errorf("failed to confirm receipt: %w", err)
	}
	return nil
}

func ChainedGet[T entities.Entity](r *relations.Relation[T], id int64, chain DatabaseChain) (record T, err error) {
	rec, dirty, err := r.Get(id)
	if err != nil {
		return
	}
	if !dirty {
		return rec, nil
	}
	synced, err := chain.IsSynced(rec.ToDatalinkRecord())
	if err != nil {
		return
	}
	if synced {
		return rec, nil
	}
	err = errors.New("no such record")
	return
}
