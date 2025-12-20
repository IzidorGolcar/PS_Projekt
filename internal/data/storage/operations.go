package storage

import (
	"context"
	"errors"
	"seminarska/internal/data/storage/db"
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"
	"time"
)

func (d *AppDatabase) GetUser(id int64) (*entities.User, error) {
	return d.Users().Get(id)
}

func (d *AppDatabase) GetMessage(id int64) (*entities.Message, error) {
	return d.Messages().Get(id)
}

func (d *AppDatabase) GetMessages(fromId int64, limit int32) ([]*entities.Message, error) {
	return d.Messages().GetPredicate(func(message *entities.Message) bool {
		return message.Id() >= fromId
	}, int(limit))
}

func (d *AppDatabase) GetLikes(messageId int64) (int, error) {
	likes, err := d.Likes().GetPredicate(func(like *entities.Like) bool {
		return like.MessageId == messageId
	}, db.NoLimit)
	if err != nil {
		return 0, err
	}
	return len(likes), nil
}

func (d *AppDatabase) GetTopics() ([]*entities.Topic, error) {
	return d.Topics().GetPredicate(func(*entities.Topic) bool {
		return true
	}, db.NoLimit)
}

func (d *AppDatabase) CreateUser(ctx context.Context, username string) (*entities.User, error) {
	user := entities.NewUser(username)
	requestId := d.chain.DispatchNewMessage(user, datalink.Operation_Create)
	id, err := d.chain.AwaitConfirmation(ctx, requestId)
	if err != nil {
		return nil, err
	}
	return d.Users().Get(id)
}

func (d *AppDatabase) CreateTopic(ctx context.Context, name string) (*entities.Topic, error) {
	user := entities.NewTopic(name)
	requestId := d.chain.DispatchNewMessage(user, datalink.Operation_Create)
	id, err := d.chain.AwaitConfirmation(ctx, requestId)
	if err != nil {
		return nil, err
	}
	return d.Topics().Get(id)
}

func (d *AppDatabase) LikeMessage(ctx context.Context, userId, messageId int64) error {
	like := entities.NewLike(messageId, userId)
	requestId := d.chain.DispatchNewMessage(like, datalink.Operation_Create)
	_, err := d.chain.AwaitConfirmation(ctx, requestId)
	return err
}

func (d *AppDatabase) PostMessage(ctx context.Context, userId, topicId int64, text string) (*entities.Message, error) {
	msg := entities.NewMessage(topicId, userId, text, time.Now())
	requestId := d.chain.DispatchNewMessage(msg, datalink.Operation_Create)
	id, err := d.chain.AwaitConfirmation(ctx, requestId)
	if err != nil {
		return nil, err
	}
	msg.SetId(id)
	return msg, nil
}

func (d *AppDatabase) DeleteMessage(ctx context.Context, userId, messageId int64) error {
	msg := entities.NewMessage(0, userId, "", time.Time{}) // dummy values
	msg.SetId(messageId)
	requestId := d.chain.DispatchNewMessage(msg, datalink.Operation_Delete)
	_, err := d.chain.AwaitConfirmation(ctx, requestId)
	return err
}

func (d *AppDatabase) UpdateMessage(ctx context.Context, userId, messageId int64, newText string) (*entities.Message, error) {
	updated, err := d.Messages().GetTransform(messageId, func(og *entities.Message) (*entities.Message, error) {
		if og.UserId != userId {
			return nil, errors.New("user mismatch")
		}
		msg := entities.NewMessage(og.TopicId, og.UserId, newText, og.CreatedAt)
		msg.SetId(og.Id())
		return msg, nil
	})
	if err != nil {
		return nil, err
	}
	requestId := d.chain.DispatchNewMessage(updated, datalink.Operation_Update)
	id, err := d.chain.AwaitConfirmation(ctx, requestId)
	if err != nil {
		return nil, err
	}
	return d.Messages().Get(id)
}

func (d *AppDatabase) SubscribeTopic(ctx context.Context, topics []int64) <-chan *entities.Message {
	out := make(chan *entities.Message, 100)
	go func() {
		defer close(out)
		for dl := range d.chain.Observe(ctx) {
			e, err := entities.DatalinkToEntity(dl)
			if err != nil {
				continue
			}
			if msg, ok := e.(*entities.Message); ok &&
				(dl.Op == datalink.Operation_Create ||
					dl.Op == datalink.Operation_Update) {

				for _, t := range topics {
					if t == msg.TopicId {
						out <- msg
					}
				}
			}
		}
	}()
	return out
}
