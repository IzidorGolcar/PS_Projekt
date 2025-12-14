package entities

type Topic struct {
	baseEntity
	name string `db:"unique"`
}

func NewTopic(name string) *Topic {
	return &Topic{name: name}
}
