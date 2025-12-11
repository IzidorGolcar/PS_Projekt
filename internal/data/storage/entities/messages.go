package entities

import (
	"seminarska/proto/datalink"
	"seminarska/proto/razpravljalnica"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Message struct {
	baseEntity
	topicId   int64     `storage:"pk"`
	userId    int64     `storage:"pk"`
	text      string    `storage:"pk"`
	createdAt time.Time `storage:"pk"`
	likes     int
}

func NewMessage(topicId int64, userId int64, text string, createdAt time.Time) *Message {
	return &Message{topicId: topicId, userId: userId, text: text, createdAt: createdAt}
}

func (r *Message) ToDatalinkRecord() *datalink.Record {
	return &datalink.Record{
		Payload: &datalink.Record_Message{Message: &razpravljalnica.Message{
			Id:        r.id,
			TopicId:   r.topicId,
			UserId:    r.userId,
			Text:      r.text,
			CreatedAt: timestamppb.New(r.createdAt),
			Likes:     int32(r.likes),
		}},
	}
}
