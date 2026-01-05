package replication

import (
	"context"
	"seminarska/internal/data/storage/db"
	"seminarska/internal/data/storage/entities"
	"seminarska/internal/data/storage/replication/broadcast"
	"seminarska/proto/datalink"
	"sync"
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
	relations             Relations
	mx                    sync.Mutex
	pendingRequests       map[int32]db.Receipt
	newMessages           chan *datalink.Message
	confirmationBroadcast *broadcast.Broadcaster[response]
	messageBroadcast      *broadcast.Broadcaster[*datalink.Message]
}

func NewHandler(relations Relations) *Handler {
	return &Handler{
		relations:             relations,
		mx:                    sync.Mutex{},
		confirmationBroadcast: broadcast.New[response](),
		messageBroadcast:      broadcast.New[*datalink.Message](),
		pendingRequests:       make(map[int32]db.Receipt),
		newMessages:           make(chan *datalink.Message),
	}
}

func (h *Handler) AwaitConfirmation(ctx context.Context, requestId string) (int64, error) {
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	for res := range h.confirmationBroadcast.Subscribe(subCtx) {
		if res.requestId == requestId {
			return res.entityId, res.err
		}
	}
	return 0, subCtx.Err()
}

func (h *Handler) Observe(ctx context.Context) <-chan *datalink.Message {
	mx := &sync.Mutex{}
	messages := make(map[string]*datalink.Message)
	out := make(chan *datalink.Message, 100)

	go func() {
		for msg := range h.messageBroadcast.Subscribe(ctx) {
			mx.Lock()
			messages[msg.GetRequestId()] = msg
			mx.Unlock()
		}
	}()

	go func() {
		defer close(out)
		for res := range h.confirmationBroadcast.Subscribe(ctx) {
			mx.Lock()
			if msg, ok := messages[res.requestId]; ok {
				delete(messages, res.requestId)
				out <- msg
			}
			mx.Unlock()
		}
	}()

	return out
}
