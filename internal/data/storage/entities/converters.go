package entities

import (
	"errors"
	"seminarska/proto/datalink"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func DatalinkToEntity(dl *datalink.Message) (entity Entity, err error) {
	switch p := dl.Payload.(type) {
	case *datalink.Message_User:
		entity = NewUser(p.User.Name)
		entity.SetId(p.User.Id)
	case *datalink.Message_Message:
		entity = NewMessage(
			p.Message.Id,
			p.Message.TopicId, p.Message.UserId,
			p.Message.Text, p.Message.CreatedAt.AsTime(),
		)
	case *datalink.Message_Like:
		entity = NewLike(p.Like.MessageId, p.Like.UserId)
		entity.SetId(p.Like.Id)
	case *datalink.Message_Topic:
		entity = NewTopic(p.Topic.Name)
		entity.SetId(p.Topic.Id)
	default:
		return nil, errors.New("invalid payload")
	}
	return
}

func EntityToDatalink(entity Entity) (dl *datalink.Message) {
	switch e := entity.(type) {
	case *User:
		return &datalink.Message{
			Payload: &datalink.Message_User{User: &razpravljalnica.User{
				Id:   e.id,
				Name: e.name,
			}},
		}
	case *Message:
		return &datalink.Message{
			Payload: &datalink.Message_Message{Message: &razpravljalnica.Message{
				Id:        e.id,
				TopicId:   e.topicId,
				UserId:    e.userId,
				Text:      e.text,
				CreatedAt: timestamppb.New(e.createdAt),
			}},
		}
	case *Topic:
		return &datalink.Message{
			Payload: &datalink.Message_Topic{Topic: &razpravljalnica.Topic{
				Id:   e.id,
				Name: e.name,
			}},
		}
	case *Like:
		return &datalink.Message{
			Payload: &datalink.Message_Like{Like: &razpravljalnica.Like{
				MessageId: e.messageId,
				UserId:    e.userId,
			}},
		}
	default:
		panic("illegal state")
	}
}
