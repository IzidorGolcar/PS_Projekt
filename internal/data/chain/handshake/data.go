package handshake

import "seminarska/proto/datalink"

type NodeState interface {
	LastMessageIndex() int32
	LastConfirmationIndex() int32
	GetMessagesAfter(int32) []*datalink.Message
	GetConfirmationsAfter(int32) []*datalink.Confirmation
	CopyDb()
}
