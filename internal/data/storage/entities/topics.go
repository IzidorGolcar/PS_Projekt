package entities

type Topic struct {
	baseEntity
	Name string `db:"unique"`
}

func NewTopic(name string) *Topic {
	return &Topic{Name: name}
}
