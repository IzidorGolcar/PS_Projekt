package storage

import (
	"errors"
	"seminarska/internal/data/storage/db"
	"seminarska/internal/data/storage/entities"
)

func (d *AppDatabase) Insert(record entities.Entity) error {
	switch r := record.(type) {
	case *entities.User:
		return ChainedInsert(d.users, r, d.chain)
	case *entities.Message:
		return ChainedInsert(d.messages, r, d.chain)
	case *entities.Topic:
		return ChainedInsert(d.topics, r, d.chain)
	case *entities.Like:
		return ChainedInsert(d.likes, r, d.chain)
	default:
		return errors.New("invalid record type")
	}
}

func (d *AppDatabase) DeleteUser(id int64) error {
	return ChainedDelete(d.users, id, d.chain)
}

func (d *AppDatabase) DeleteMessage(id int64) error {
	return ChainedDelete(d.messages, id, d.chain)
}

func (d *AppDatabase) UpdateMessage(id int64, transform db.TransformFunc[*entities.Message]) error {
	return ChainedUpdate(d.messages, id, transform, d.chain)
}
