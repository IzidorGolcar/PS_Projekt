package requests

import (
	"context"
	"seminarska/internal/data/storage"
	"seminarska/proto/razpravljalnica"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (l *listener) CreateUser(
	ctx context.Context,
	request *razpravljalnica.CreateUserRequest,
) (*razpravljalnica.User, error) {
	user := storage.NewUserRecord(request.GetName())
	err := l.db.Insert(user)
	if err != nil {
		return nil, err
	}
	return user.ToDatalinkRecord().GetUser(), nil
}

func (l *listener) CreateTopic(
	ctx context.Context,
	request *razpravljalnica.CreateTopicRequest,
) (*razpravljalnica.Topic, error) {
	topic := storage.NewTopicRecord(request.GetName())
	err := l.db.Insert(topic)
	if err != nil {
		return nil, err
	}
	return topic.ToDatalinkRecord().GetTopic(), nil
}

func (l *listener) PostMessage(
	ctx context.Context,
	request *razpravljalnica.PostMessageRequest,
) (*razpravljalnica.Message, error) {
	msg := storage.NewMessageRecord(
		request.GetTopicId(), request.GetUserId(),
		request.GetText(), time.Now(),
	)
	err := l.db.Insert(msg)
	if err != nil {
		return nil, err
	}
	return msg.ToDatalinkRecord().GetMessage(), nil
}

func (l *listener) UpdateMessage(
	ctx context.Context,
	request *razpravljalnica.UpdateMessageRequest,
) (*razpravljalnica.Message, error) {
	//TODO implement me
	panic("implement me")
}

func (l *listener) DeleteMessage(ctx context.Context, request *razpravljalnica.DeleteMessageRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (l *listener) LikeMessage(ctx context.Context, request *razpravljalnica.LikeMessageRequest) (*razpravljalnica.Message, error) {
	//TODO implement me
	panic("implement me")
}

func (l *listener) GetSubcscriptionNode(ctx context.Context, request *razpravljalnica.SubscriptionNodeRequest) (*razpravljalnica.SubscriptionNodeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *listener) ListTopics(ctx context.Context, empty *emptypb.Empty) (*razpravljalnica.ListTopicsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *listener) GetMessages(ctx context.Context, request *razpravljalnica.GetMessagesRequest) (*razpravljalnica.GetMessagesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *listener) SubscribeTopic(request *razpravljalnica.SubscribeTopicRequest, g grpc.ServerStreamingServer[razpravljalnica.MessageEvent]) error {
	//TODO implement me
	panic("implement me")
}
