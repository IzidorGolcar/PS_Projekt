package storage

type TopicRecord struct {
	baseRecord
	name string
}

func NewTopicRecord(name string) TopicRecord {
	return TopicRecord{name: name}
}
