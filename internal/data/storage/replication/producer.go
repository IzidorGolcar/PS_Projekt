package replication

import (
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"
)

func (h *Handler) Messages() <-chan *datalink.Message {
	return h.newMessages
}

func (h *Handler) DispatchNewMessage(entity entities.Entity, operation datalink.Operation) {
	message := entities.EntityToDatalink(entity)
	message.Op = operation
	h.newMessages <- message
}
