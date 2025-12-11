package storage

import (
	"errors"
	"seminarska/internal/data/storage/entities"
)

func (d *Database) Insert(record entities.Entity) error {
	switch r := record.(type) {
	case *entities.User:
		return ChainedInsert[*entities.User](d.users, r, d.chain)
	case *entities.Message:
		return ChainedInsert[*entities.Message](d.messages, r, d.chain)
	case *entities.Topic:
		return ChainedInsert[*entities.Topic](d.topics, r, d.chain)
	default:
		return errors.New("invalid record type")
	}
}

func (d *Database) GetUser(id int64) (rec *entities.User, err error) {
	return ChainedGet[*entities.User](d.users, id, d.chain)
}

func (d *Database) GetMessage(id int64) (rec *entities.Message, err error) {
	return ChainedGet[*entities.Message](d.messages, id, d.chain)
}

func (d *Database) GetTopic(id int64) (rec *entities.Topic, err error) {
	return ChainedGet[*entities.Topic](d.topics, id, d.chain)
}
