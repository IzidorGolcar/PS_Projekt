package storage

import (
	"context"
	"seminarska/internal/data/storage/db"
	"seminarska/internal/data/storage/entities"
)

type AppDatabase struct {
	ctx      context.Context
	cancel   context.CancelFunc
	done     chan struct{}
	chain    ChainHandler
	users    *db.Relation[*entities.User]
	messages *db.Relation[*entities.Message]
	topics   *db.Relation[*entities.Topic]
	likes    *db.Relation[*entities.Like]
}

func NewAppDatabase(ctx context.Context, chain ChainHandler) *AppDatabase {
	dbCtx, cancel := context.WithCancel(ctx)
	return &AppDatabase{
		ctx:      dbCtx,
		cancel:   cancel,
		done:     make(chan struct{}),
		chain:    chain,
		users:    db.NewRelation[*entities.User](),
		messages: db.NewRelation[*entities.Message](),
		topics:   db.NewRelation[*entities.Topic](),
		likes:    db.NewRelation[*entities.Like](),
	}
}

func (d *AppDatabase) Close() {
	d.cancel()
}

func (d *AppDatabase) Done() <-chan struct{} {
	return d.done
}
