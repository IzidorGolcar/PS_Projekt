package entities

import (
	"time"
)

type Message struct {
	baseEntity
	TopicId   int64 `db:"unique"`
	UserId    int64 `db:"unique"`
	Text      string
	CreatedAt time.Time `db:"unique"`
}

func NewMessage(topicId int64, userId int64, text string, createdAt time.Time) *Message {
	return &Message{TopicId: topicId, UserId: userId, Text: text, CreatedAt: createdAt}
}
