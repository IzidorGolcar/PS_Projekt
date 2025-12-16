package storage

import (
	"seminarska/internal/data/storage/db"
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"
	"time"
)

func (d *AppDatabase) GetMessages(fromId int64, limit int32) ([]*entities.Message, error) {
	return d.Messages().GetPredicate(func(message *entities.Message) bool {
		return message.Id() >= fromId
	}, int(limit))
}

func (d *AppDatabase) GetLikes(messageId int64) (int, error) {
	likes, err := d.Likes().GetPredicate(func(like *entities.Like) bool {
		return like.MessageId() == messageId
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

func (d *AppDatabase) CreateUser(username string) error {
	user := entities.NewUser(username)
	d.chain.DispatchNewMessage(user, datalink.Operation_Create)
	panic("todo: wait for confirmation")
}

func (d *AppDatabase) UpdateMessage(userId, messageId int64, newText string) error {
	message := entities.NewMessage(0, userId, newText, time.Time{})
	message.SetId(messageId)
	d.chain.DispatchNewMessage(message, datalink.Operation_Update)
	// fixme: this would overwrite the old message.
	panic("todo: wait for confirmation")
}
