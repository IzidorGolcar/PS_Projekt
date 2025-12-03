package storage

type TopicRecord struct {
	id   int64
	name string
}

func NewTopicRecord(name string) *TopicRecord {
	return &TopicRecord{name: name}
}

func (t TopicRecord) hash() uint64 {
	return stringHash(t.name)
}

func (t TopicRecord) Id() RecordId {
	return RecordId(t.id)
}

type Topics struct {
	*defaultRelation[TopicRecord]
}

func NewTopics() *Topics {
	return &Topics{newDefaultRelation[TopicRecord]()}
}
