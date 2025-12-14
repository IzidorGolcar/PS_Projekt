package storage

import (
	"errors"
	"seminarska/internal/data/storage/db"
	"seminarska/internal/data/storage/entities"
)

func (d *AppDatabase) Insert(record entities.Entity) error {
	switch r := record.(type) {
	case *entities.User:
		return ChainedInsert(d.users, r, d.chain)
	case *entities.Message:
		return ChainedInsert(d.messages, r, d.chain)
	case *entities.Topic:
		return ChainedInsert(d.topics, r, d.chain)
	default:
		return errors.New("invalid record type")
	}
}

func (d *AppDatabase) DeleteUser(id int64) error {
	return ChainedDelete(d.users, id, d.chain)
}

func (d *AppDatabase) DeleteMessage(id int64) error {
	return ChainedDelete(d.messages, id, d.chain)
}

func (d *AppDatabase) UpdateMessage(id int64, transform db.TransformFunc[*entities.Message]) error {
	return ChainedUpdate(d.messages, id, transform, d.chain)
}

func (d *AppDatabase) GetUser(id int64) (*entities.User, error) {
	return ChainedGet(d.users, id, d.chain)
}

func (d *AppDatabase) GetTopic(id int64) (*entities.Topic, error) {
	return ChainedGet(d.topics, id, d.chain)
}

func (d *AppDatabase) GetMessage(id int64) (*entities.Message, error) {
	return ChainedGet(d.messages, id, d.chain)
}

func (d *AppDatabase) GetLike(id int64) (*entities.Like, error) {
	return ChainedGet(d.likes, id, d.chain)
}

func (d *AppDatabase) GetMessages(fromId int64, limit int) (rec []*entities.Message, err error) {
	return ChainedGetPredicate(d.messages, d.chain, func(message *entities.Message) bool {
		return fromId == 0 || message.Id() >= fromId
	}, limit)
}

func (d *AppDatabase) GetTopics(fromId int64, limit int) (rec []*entities.Topic, err error) {
	return ChainedGetPredicate(d.topics, d.chain, func(topic *entities.Topic) bool {
		return fromId == 0 || topic.Id() >= fromId
	}, limit)
}

func (d *AppDatabase) LikeMessage(messageId, userId int64) error {
	return ChainedInsert(d.likes, entities.NewLike(userId, messageId), d.chain)
}

func (d *AppDatabase) GetLikes(messageId int64) ([]*entities.Like, error) {
	return ChainedGetPredicate(d.likes, d.chain, func(like *entities.Like) bool {
		return like.MessageId() == messageId
	}, 0)
}
