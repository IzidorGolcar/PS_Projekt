package entities

import (
	"seminarska/proto/datalink"
	"seminarska/proto/razpravljalnica"
)

type User struct {
	baseEntity
	name string `db:"unique"`
}

func NewUser(name string) *User {
	return &User{name: name}
}

func (r *User) ToDatalinkRecord() *datalink.Record {
	return &datalink.Record{
		Payload: &datalink.Record_User{User: &razpravljalnica.User{
			Id:   r.id,
			Name: r.name,
		}},
	}
}
