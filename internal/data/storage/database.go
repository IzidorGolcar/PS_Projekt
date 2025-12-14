package storage

import (
	"seminarska/internal/data/storage/db"
	"seminarska/internal/data/storage/entities"
)

type AppDatabase struct {
	chain    ChainHandler
	users    *db.Relation[*entities.User]
	messages *db.Relation[*entities.Message]
	topics   *db.Relation[*entities.Topic]
	likes    *db.Relation[*entities.Like]
}

func NewAppDatabase(chain ChainHandler) *AppDatabase {
	return &AppDatabase{
		chain:    chain,
		users:    db.NewRelation[*entities.User](),
		messages: db.NewRelation[*entities.Message](),
		topics:   db.NewRelation[*entities.Topic](),
		likes:    db.NewRelation[*entities.Like](),
	}
}
