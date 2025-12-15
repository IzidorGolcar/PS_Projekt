package entities

type Topic struct {
	baseEntity
	name string `db:"unique"`
}

func (t *Topic) Name() string {
	return t.name
}

func NewTopic(name string) *Topic {
	return &Topic{name: name}
}
