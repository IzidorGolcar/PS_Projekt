package handshake

import "seminarska/proto/datalink"

type ServerData interface {
	LastMessageIndex() int32
	GetConfirmationsAfter(int32) []*datalink.Confirmation
	ProcessMessages([]*datalink.Message)
	DatabaseImporter
}

type ClientData interface {
	LastConfirmationIndex() int32
	GetMessagesAfter(int32) []*datalink.Message
	ProcessConfirmations([]*datalink.Confirmation)
	DatabaseExporter
}

type DatabaseExporter interface {
	GetSnapshot() *datalink.DatabaseSnapshot
}

type DatabaseImporter interface {
	SetFromSnapshot(snapshot *datalink.DatabaseSnapshot)
}

type DatabaseTransfer interface {
	DatabaseExporter
	DatabaseImporter
}
