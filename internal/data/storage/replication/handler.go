package replication

import (
	"context"
	"seminarska/internal/common/broadcast"
	"seminarska/internal/data/storage/db"
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"
)

type Relations interface {
	Users() *db.Relation[*entities.User]
	Messages() *db.Relation[*entities.Message]
	Topics() *db.Relation[*entities.Topic]
	Likes() *db.Relation[*entities.Like]
}

type response struct {
	requestId string
	entityId  int64
	err       error
}

func newResponse(requestId string, entityId int64, err error) response {
	return response{requestId: requestId, entityId: entityId, err: err}
}

type Handler struct {
	relations       Relations
	pendingRequests map[int32]db.Receipt
	newMessages     chan *datalink.Message
	broadcast       *broadcast.Broadcaster[response]
}

func NewHandler(relations Relations) *Handler {
	return &Handler{
		relations:       relations,
		broadcast:       broadcast.New[response](),
		pendingRequests: make(map[int32]db.Receipt),
		newMessages:     make(chan *datalink.Message),
	}
}

func (h *Handler) AwaitConfirmation(ctx context.Context, requestId string) (int64, error) {
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	for res := range h.broadcast.Subscribe(subCtx) {
		if res.requestId == requestId {
			return res.entityId, res.err
		}
	}
	return 0, subCtx.Err()
}
