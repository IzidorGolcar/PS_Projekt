package entities

type Topic struct {
	BaseEntity
	Name string `db:"unique"`
}

func NewTopic(name string) *Topic {
	return &Topic{Name: name}
}
