package storage

import "time"

type MessageRecord struct {
	id        int64
	topicId   int64
	userId    int64
	text      string
	createdAt time.Time
	likes     int
}

func NewMessageRecord(topicId int64, userId int64, text string, createdAt time.Time) *MessageRecord {
	return &MessageRecord{topicId: topicId, userId: userId, text: text, createdAt: createdAt}
}

func (m MessageRecord) hash() uint64 {
	return stringHash(m.text, m.userId, m.topicId, m.createdAt)
}

func (m MessageRecord) Id() RecordId {
	return RecordId(m.id)
}

type Messages struct {
	*defaultRelation[MessageRecord]
}

func NewMessages() *Messages {
	return &Messages{newDefaultRelation[MessageRecord]()}
}
