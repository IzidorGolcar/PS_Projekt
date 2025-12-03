package storage

import (
	"errors"
)

type RecordId int64

type Record interface {
	hash() uint64
	Id() RecordId
}

type Receipt interface {
	Confirm(id RecordId) error
	Cancel(err error)
}

type relation[T Record] interface {
	Insert(T) (Receipt, error)
	Get(id RecordId) (T, error)
}

type Database struct {
	users *Users
}

func NewDatabase() *Database {
	return &Database{
		users: NewUsers(),
	}
}

func (d *Database) Insert(record Record) (rec Receipt, err error) {
	switch r := record.(type) {
	case *UserRecord:
		rec, err = d.users.Insert(*r)
	default:
		err = errors.New("invalid record type")
	}
	return
}

func (d *Database) GetUser(id RecordId) (rec UserRecord, err error) {
	return d.users.Get(id)
}
