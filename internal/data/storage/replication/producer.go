package replication

import (
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"

	"github.com/google/uuid"
)

func (h *Handler) Messages() <-chan *datalink.Message {
	return h.newMessages
}

func (h *Handler) DispatchNewMessage(entity entities.Entity, operation datalink.Operation) string {
	requestId := uuid.New().String()
	message := entities.EntityToDatalink(entity)
	message.RequestId = requestId
	message.Op = operation
	h.newMessages <- message
	return requestId
}
