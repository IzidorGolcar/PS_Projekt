package storage

import (
	"errors"
	"fmt"
	"seminarska/internal/data/storage/entities"
	"seminarska/internal/data/storage/relations"
	"seminarska/proto/datalink"
)

type Chain interface {
	ForwardRequest(*datalink.Record) error
	SyncedWithTail(*datalink.Record) (bool, error)
}

func ChainedInsert[T entities.Entity](r *relations.Relation[T], record T, chain Chain) error {
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

func ChainedGet[T entities.Entity](r *relations.Relation[T], id int64, chain Chain) (record T, err error) {
	var res relations.Result[T]
	res, err = r.Get(id)
	if err != nil {
		return
	}
	if res.confirmed {
		return res.record, nil
	}
	synced, err := chain.SyncedWithTail(res.record.ToDatalinkRecord())
	if err != nil {
		return
	}
	if synced {
		return res.record, nil
	}
	err = errors.New("no such record")
	return
}
