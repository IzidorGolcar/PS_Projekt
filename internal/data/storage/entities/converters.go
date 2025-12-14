package entities

import (
	"errors"
	"seminarska/proto/datalink"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func DatalinkToEntity(dl *datalink.Record) (entity Entity, err error) {
	switch p := dl.Payload.(type) {
	case *datalink.Record_User:
		entity = NewUser(p.User.Name)
		entity.SetId(p.User.Id)
	case *datalink.Record_Message:
		entity = NewMessage(
			p.Message.TopicId, p.Message.UserId,
			p.Message.Text, p.Message.CreatedAt.AsTime(),
		)
		entity.SetId(p.Message.Id)
	case *datalink.Record_Like:
		entity = NewLike(p.Like.MessageId, p.Like.UserId)
		entity.SetId(p.Like.Id)
	case *datalink.Record_Topic:
		entity = NewTopic(p.Topic.Name)
		entity.SetId(p.Topic.Id)
	default:
		return nil, errors.New("invalid payload")
	}
	return
}

func EntityToDatalink(entity Entity) (dl *datalink.Record) {
	switch e := entity.(type) {
	case *User:
		return &datalink.Record{
			Payload: &datalink.Record_User{User: &razpravljalnica.User{
				Id:   e.id,
				Name: e.name,
			}},
		}
	case *Message:
		return &datalink.Record{
			Payload: &datalink.Record_Message{Message: &razpravljalnica.Message{
				Id:        e.id,
				TopicId:   e.topicId,
				UserId:    e.userId,
				Text:      e.text,
				CreatedAt: timestamppb.New(e.createdAt),
			}},
		}
	case *Topic:
		return &datalink.Record{
			Payload: &datalink.Record_Topic{Topic: &razpravljalnica.Topic{
				Id:   e.id,
				Name: e.name,
			}},
		}
	case *Like:
		return &datalink.Record{
			Payload: &datalink.Record_Like{Like: &razpravljalnica.Like{
				MessageId: e.messageId,
				UserId:    e.userId,
			}},
		}
	default:
		panic("illegal state")
	}
}
