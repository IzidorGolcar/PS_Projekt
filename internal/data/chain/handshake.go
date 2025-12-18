package chain

import "seminarska/proto/datalink"

type Observer struct {
	messages      *ReplayBuffer[*datalink.Message]
	confirmations *ReplayBuffer[*datalink.Confirmation]
}

func (o *Observer) LastMessageIndex() int32 {
	i, err := o.messages.LastMessageIndex()
	if err != nil {
		return -1
	}
	return i
}

func (o *Observer) LastConfirmationIndex() int32 {
	i, err := o.confirmations.LastMessageIndex()
	if err != nil {
		return -1
	}
	return i
}

func (o *Observer) GetMessagesAfter(i int32) ([]*datalink.Message, error) {
	return o.messages.MessagesAfter(i), nil
}

func (o *Observer) GetConfirmationsAfter(i int32) ([]*datalink.Confirmation, error) {
	return o.confirmations.MessagesAfter(i), nil
}
