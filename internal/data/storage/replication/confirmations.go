package replication

import (
	"errors"
	"log"
	"seminarska/proto/datalink"
)

func (h *Handler) OnConfirmation(confirmation *datalink.Confirmation) {
	h.broadcast.Broadcast(newResponse(
		confirmation.GetRequestId(), int64(confirmation.GetMessageIndex()),
	))
	receipt, ok := h.pendingRequests[confirmation.GetMessageIndex()]
	if !ok {
		return
	}
	if confirmation.Ok {
		err := receipt.Confirm()
		if err != nil {
			log.Println("Failed to confirm record", err)
		}
	} else {
		receipt.Cancel(errors.New(confirmation.GetError()))
	}
}
