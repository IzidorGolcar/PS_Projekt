package entities

import (
	"seminarska/proto/datalink"
	"seminarska/proto/razpravljalnica"
)

type Topic struct {
	baseEntity
	name string `storage:"pk"`
}

func NewTopic(name string) *Topic {
	return &Topic{name: name}
}

func (r *Topic) ToDatalinkRecord() *datalink.Record {
	return &datalink.Record{
		Payload: &datalink.Record_Topic{Topic: &razpravljalnica.Topic{
			Id:   int64(r.id),
			Name: r.name,
		}},
	}
}
