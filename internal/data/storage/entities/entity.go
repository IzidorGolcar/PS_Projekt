package entities

type Entity interface {
	SetId(id int64)
	Id() int64
}

type BaseEntity struct {
	id int64
}

func (b *BaseEntity) SetId(id int64) {
	b.id = id
}

func (b *BaseEntity) Id() int64 {
	return b.id
}
