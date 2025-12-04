package storage

import "time"

type MessageRecord struct {
	baseRecord
	topicId   int64
	userId    int64
	text      string
	createdAt time.Time
	likes     int
}

func NewMessageRecord(topicId int64, userId int64, text string, createdAt time.Time) MessageRecord {
	return MessageRecord{topicId: topicId, userId: userId, text: text, createdAt: createdAt}
}
