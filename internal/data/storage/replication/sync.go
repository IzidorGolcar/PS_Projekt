package replication

import (
	"errors"
	"seminarska/internal/data/storage/db"
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"
)

func (h *Handler) OnMessage(message *datalink.Message) error {
	h.messageBroadcast.Broadcast(message)
	entity, err := entities.DatalinkToEntity(message)
	if err != nil {
		return err
	}
	entity.SetId(int64(message.MessageIndex))
	receipt, err := h.chainedOperation(entity, message.GetOp())
	if err != nil {
		return err
	}
	h.pendingRequests[message.GetMessageIndex()] = receipt
	return nil
}

func (h *Handler) chainedOperation(entity entities.Entity, op datalink.Operation) (rec db.Receipt, err error) {
	switch v := entity.(type) {
	case *entities.Message:
		rec, err = do(h.relations.Messages(), v, op)
	case *entities.User:
		rec, err = do(h.relations.Users(), v, op)
	case *entities.Like:
		rec, err = do(h.relations.Likes(), v, op)
	case *entities.Topic:
		rec, err = do(h.relations.Topics(), v, op)
	default:
		err = errors.New("invalid entity")
	}
	return
}

func do[E entities.Entity](relation *db.Relation[E], entity E, op datalink.Operation) (db.Receipt, error) {
	switch op {
	case datalink.Operation_Create:
		return relation.Insert(entity)
	case datalink.Operation_Delete:
		return relation.Delete(entity.Id())
	case datalink.Operation_Update:
		return relation.Update(entity.Id(), func(e E) (E, error) { return entity, nil })
	default:
		return nil, errors.New("invalid operation")
	}
}
