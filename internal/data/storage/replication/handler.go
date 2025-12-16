package replication

import (
	"seminarska/internal/data/storage/db"
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"
	"sync"
)

type Relations interface {
	Users() *db.Relation[*entities.User]
	Messages() *db.Relation[*entities.Message]
	Topics() *db.Relation[*entities.Topic]
	Likes() *db.Relation[*entities.Like]
}

type Handler struct {
	relations       Relations
	pendingRequests map[int32]db.Receipt
	newMessages     chan *datalink.Message
	mx              *sync.RWMutex
}

func NewHandler(relations Relations) *Handler {
	return &Handler{
		relations:       relations,
		pendingRequests: make(map[int32]db.Receipt),
		newMessages:     make(chan *datalink.Message),
	}
}
