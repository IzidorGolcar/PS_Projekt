package entities

import (
	"seminarska/proto/datalink"
)

type Entity interface {
	ToDatalinkRecord() *datalink.Record
	SetId(id int64)
	Id() int64
}

type baseEntity struct {
	id int64
}

func (b *baseEntity) SetId(id int64) {
	b.id = id
}

func (b *baseEntity) Id() int64 {
	return b.id
}
