package storage

import (
	"seminarska/internal/data/storage/db"
	"seminarska/internal/data/storage/entities"
	"seminarska/internal/data/storage/replication"
)

type relations struct {
	users    *db.Relation[*entities.User]
	messages *db.Relation[*entities.Message]
	topics   *db.Relation[*entities.Topic]
	likes    *db.Relation[*entities.Like]
}

func (d *relations) Users() *db.Relation[*entities.User] {
	return d.users
}

func (d *relations) Messages() *db.Relation[*entities.Message] {
	return d.messages
}

func (d *relations) Topics() *db.Relation[*entities.Topic] {
	return d.topics
}

func (d *relations) Likes() *db.Relation[*entities.Like] {
	return d.likes
}

type AppDatabase struct {
	*relations
	chain *replication.Handler
}

func NewAppDatabase() *AppDatabase {
	relations := &relations{
		users:    db.NewRelation[*entities.User](),
		messages: db.NewRelation[*entities.Message](),
		topics:   db.NewRelation[*entities.Topic](),
		likes:    db.NewRelation[*entities.Like](),
	}
	return &AppDatabase{
		relations: relations,
		chain:     replication.NewHandler(relations),
	}
}

func (d *AppDatabase) Chain() *replication.Handler {
	return d.chain
}
