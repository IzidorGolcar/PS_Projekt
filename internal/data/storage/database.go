package storage

import (
	"seminarska/internal/data/storage/entities"
	"seminarska/internal/data/storage/relations"
)

type Database struct {
	chain    DatabaseChain
	users    *relations.Relation[*entities.User]
	messages *relations.Relation[*entities.Message]
	topics   *relations.Relation[*entities.Topic]
}

func NewDatabase(chain DatabaseChain) *Database {
	return &Database{
		chain:    chain,
		users:    relations.NewRelation[*entities.User](),
		messages: relations.NewRelation[*entities.Message](),
		topics:   relations.NewRelation[*entities.Topic](),
	}
}
