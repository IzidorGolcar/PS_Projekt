package storage

import "errors"

func (d *Database) Insert(record Record) (rec Receipt, err error) {
	switch r := record.(type) {
	case *UserRecord:
		rec, err = d.users.Insert(r)
	case *MessageRecord:
		rec, err = d.messages.Insert(r)
	case *TopicRecord:
		rec, err = d.topics.Insert(r)
	default:
		err = errors.New("invalid record type")
	}
	return
}

func (d *Database) GetUser(id uint64) (rec *UserRecord, err error) {
	return d.users.Get(id)
}
