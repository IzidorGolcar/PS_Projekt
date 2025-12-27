package replication

import (
	"errors"
	"log"
	"seminarska/proto/datalink"
)

func (h *Handler) OnConfirmation(confirmation *datalink.Confirmation) {
	var err error
	if !confirmation.Ok {
		err = errors.New(confirmation.GetError())
	}
	h.confirmationBroadcast.Broadcast(newResponse(
		confirmation.GetRequestId(),
		int64(confirmation.GetMessageIndex()),
		err,
	))
	h.mx.Lock()
	receipt, ok := h.pendingRequests[confirmation.GetMessageIndex()]
	if !ok {
		h.mx.Unlock()
		return
	}
	delete(h.pendingRequests, confirmation.GetMessageIndex())
	h.mx.Unlock()
	if confirmation.Ok {
		err := receipt.Confirm()
		if err != nil {
			log.Println("Failed to confirm record", err)
		}
	} else {
		receipt.Cancel(errors.New(confirmation.GetError()))
	}
}
