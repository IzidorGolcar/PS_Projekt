package requests

import (
	"context"
	"errors"
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (l *listener) GetUser(_ context.Context, req *razpravljalnica.GetUserRequest) (*razpravljalnica.User, error) {
	if req.UserId == nil && req.Username == nil {
		return nil, errors.New("bad request")
	}
	var user *entities.User
	var err error
	if req.UserId != nil {
		user, err = l.db.Users().Get(req.GetUserId())
	} else {
		var users []*entities.User
		users, err = l.db.Users().GetPredicate(func(user *entities.User) bool {
			return user.Name == req.GetUsername()
		}, 1)
		if err == nil {
			if len(users) == 1 {
				user = users[0]
			} else {
				err = errors.New("not found")
			}
		}
	}
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
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
	messages, err := l.db.GetMessages(request.GetFromMessageId(), request.GetTopicId(), request.GetLimit())
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

func (l *listener) SubscribeTopic(
	request *razpravljalnica.SubscribeTopicRequest,
	g grpc.ServerStreamingServer[razpravljalnica.MessageEvent],
) error {
	if request.GetSubscribeToken() != l.subToken {
		return status.Error(codes.Unauthenticated, "invalid token")
	}
	for e := range l.db.SubscribeTopic(g.Context(), request.GetTopicId()) {
		var op razpravljalnica.OpType
		switch e.Operation {
		case datalink.Operation_Delete:
			op = razpravljalnica.OpType_OP_DELETE
		case datalink.Operation_Update:
			op = razpravljalnica.OpType_OP_UPDATE
		case datalink.Operation_Create:
			op = razpravljalnica.OpType_OP_POST
		}

		rMessage := &razpravljalnica.MessageEvent{
			Message: entities.EntityToDatalink(e.Message).GetMessage(),
			Op:      op,
		}
		likes, err := l.db.GetLikes(e.Message.Id())
		if err != nil {
			return err
		}
		rMessage.Message.Likes = int32(likes)
		err = g.Send(rMessage)
		if err != nil {
			return err
		}
	}
	return nil
}
