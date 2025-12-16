package requests

import (
	"context"
	"errors"
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/razpravljalnica"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (l *listener) CreateUser(
	_ context.Context,
	request *razpravljalnica.CreateUserRequest,
) (*razpravljalnica.User, error) {

}

func (l *listener) CreateTopic(
	_ context.Context,
	request *razpravljalnica.CreateTopicRequest,
) (*razpravljalnica.Topic, error) {

}

func (l *listener) PostMessage(
	_ context.Context,
	request *razpravljalnica.PostMessageRequest,
) (*razpravljalnica.Message, error) {
	timestamp := time.Now()
	message := entities.NewMessage(request.TopicId, request.UserId, request.Text, timestamp)
	err := l.db.Insert(message)
	if err != nil {
		return nil, err
	}
	return entities.EntityToDatalink(message).GetMessage(), nil
}

func (l *listener) UpdateMessage(
	_ context.Context,
	request *razpravljalnica.UpdateMessageRequest,
) (out *razpravljalnica.Message, err error) {
	err = l.db.UpdateMessage(request.MessageId, func(current *entities.Message) (*entities.Message, error) {
		if current.TopicId() != request.TopicId {
			return nil, errors.New("topic mismatch")
		}
		if current.UserId() != request.UserId {
			return nil, errors.New("user mismatch")
		}
		updated := entities.NewMessage(
			current.Id(), current.UserId(), request.Text, current.CreatedAt(),
		)
		out = entities.EntityToDatalink(updated).GetMessage()
		return updated, nil
	})
	likes, err := l.db.GetLikes(request.MessageId)
	if err != nil {
		return nil, err
	}
	out.Likes = int32(len(likes))
	return
}

func (l *listener) DeleteMessage(
	_ context.Context,
	request *razpravljalnica.DeleteMessageRequest,
) (*emptypb.Empty, error) {
	err := l.db.DeleteMessage(request.GetMessageId())
	return &emptypb.Empty{}, err
}

func (l *listener) LikeMessage(
	_ context.Context,
	request *razpravljalnica.LikeMessageRequest,
) (*razpravljalnica.Message, error) {
	err := l.db.LikeMessage(request.MessageId, request.UserId)
	if err != nil {
		return nil, err
	}
	msg, err := l.db.GetMessage(request.MessageId)
	if err != nil {
		return nil, err
	}
	return entities.EntityToDatalink(msg).GetMessage(), nil
}

func (l *listener) ListTopics(
	_ context.Context,
	_ *emptypb.Empty,
) (*razpravljalnica.ListTopicsResponse, error) {
	topics, err := l.db.GetTopics()
	if err != nil {
		return nil, err
	}
	out := make([]*razpravljalnica.Topic, len(topics))
	for i, topic := range topics {
		out[i] = entities.EntityToDatalink(topic).GetTopic()
	}
	return &razpravljalnica.ListTopicsResponse{
		Topics: out,
	}, nil
}

func (l *listener) GetMessages(
	_ context.Context,
	request *razpravljalnica.GetMessagesRequest,
) (*razpravljalnica.GetMessagesResponse, error) {
	messages, err := l.db.GetMessages(request.GetFromMessageId(), request.GetLimit())
	if err != nil {
		return nil, err
	}
	out := make([]*razpravljalnica.Message, len(messages))
	for i, message := range messages {
		msg := entities.EntityToDatalink(message).GetMessage()
		if msg == nil {
			panic("illegal state")
		}
		likes, err := l.db.GetLikes(message.Id())
		if err != nil {
			return nil, err
		}
		msg.Likes = int32(likes)
		out[i] = msg
	}
	return &razpravljalnica.GetMessagesResponse{
		Messages: out,
	}, nil
}

func (l *listener) GetSubcscriptionNode(
	_ context.Context,
	request *razpravljalnica.SubscriptionNodeRequest,
) (*razpravljalnica.SubscriptionNodeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *listener) SubscribeTopic(
	request *razpravljalnica.SubscribeTopicRequest,
	g grpc.ServerStreamingServer[razpravljalnica.MessageEvent],
) error {
	//TODO implement me
	panic("implement me")
}
