package requests

import (
	"context"
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (l *listener) CreateUser(
	ctx context.Context,
	request *razpravljalnica.CreateUserRequest,
) (*razpravljalnica.User, error) {
	user, err := l.db.CreateUser(ctx, request.GetName())
	if err != nil {
		return nil, err
	}
	return entities.EntityToDatalink(user).GetUser(), nil
}

func (l *listener) CreateTopic(
	ctx context.Context,
	request *razpravljalnica.CreateTopicRequest,
) (*razpravljalnica.Topic, error) {
	topic, err := l.db.CreateTopic(ctx, request.GetName())
	if err != nil {
		return nil, err
	}
	return entities.EntityToDatalink(topic).GetTopic(), nil
}

func (l *listener) PostMessage(
	ctx context.Context,
	request *razpravljalnica.PostMessageRequest,
) (*razpravljalnica.Message, error) {
	msg, err := l.db.PostMessage(ctx, request.GetUserId(), request.GetTopicId(), request.GetText())
	if err != nil {
		return nil, err
	}
	return entities.EntityToDatalink(msg).GetMessage(), nil
}

func (l *listener) UpdateMessage(
	ctx context.Context,
	request *razpravljalnica.UpdateMessageRequest,
) (*razpravljalnica.Message, error) {
	msg, err := l.db.UpdateMessage(ctx, request.GetUserId(), request.GetMessageId(), request.GetText())
	if err != nil {
		return nil, err
	}
	likes, err := l.db.GetLikes(request.MessageId)
	if err != nil {
		return nil, err
	}
	out := entities.EntityToDatalink(msg).GetMessage()
	if out == nil {
		panic("illegal state:")
	}
	out.Likes = int32(likes)
	return out, nil
}

func (l *listener) DeleteMessage(
	ctx context.Context,
	request *razpravljalnica.DeleteMessageRequest,
) (*emptypb.Empty, error) {
	err := l.db.DeleteMessage(ctx, request.GetUserId(), request.GetMessageId())
	return &emptypb.Empty{}, err
}

func (l *listener) LikeMessage(
	ctx context.Context,
	request *razpravljalnica.LikeMessageRequest,
) (*razpravljalnica.Message, error) {
	err := l.db.LikeMessage(ctx, request.MessageId, request.UserId)
	if err != nil {
		return nil, err
	}
	msg, err := l.db.GetMessage(request.MessageId)
	if err != nil {
		return nil, err
	}
	out := entities.EntityToDatalink(msg).GetMessage()
	if out == nil {
		panic("illegal state:")
	}
	out.Likes++
	return out, nil
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

// SubscribeTopic streams message events for a given topic to the client.
// Note: The stream parameter uses the generated MessageBoard_SubscribeTopicServer interface
// instead of the generic grpc.ServerStreamingServer[T] to match the interface expected by
// the protoc-gen-go-grpc generated code.
func (l *listener) SubscribeTopic(
	request *razpravljalnica.SubscribeTopicRequest,
	stream razpravljalnica.MessageBoard_SubscribeTopicServer,
) error {
	for msg := range l.db.SubscribeTopic(stream.Context(), request.GetTopicId()) {
		rMessage := &razpravljalnica.MessageEvent{Message: entities.EntityToDatalink(msg).GetMessage()}
		likes, err := l.db.GetLikes(msg.Id())
		if err != nil {
			return err
		}
		rMessage.Message.Likes = int32(likes)
		err = stream.Send(rMessage)
		if err != nil {
			return err
		}
	}
	return nil
}
