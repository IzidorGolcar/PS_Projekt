package entities

import (
	"time"
)

type Message struct {
	baseEntity
	topicId   int64 `db:"unique"`
	userId    int64 `db:"unique"`
	text      string
	createdAt time.Time `db:"unique"`
}

func (m *Message) TopicId() int64 {
	return m.topicId
}

func (m *Message) UserId() int64 {
	return m.userId
}

func (m *Message) Text() string {
	return m.text
}

func (m *Message) CreatedAt() time.Time {
	return m.createdAt
}

func NewMessage(topicId int64, userId int64, text string, createdAt time.Time) *Message {
	return &Message{topicId: topicId, userId: userId, text: text, createdAt: createdAt}
}
