package chain

import (
	"log"
	"seminarska/internal/data/chain/handshake"
	"seminarska/proto/datalink"
	"sync/atomic"
)

type OpCounter struct {
	n int32
}

func NewOpCounter(initial int32) *OpCounter {
	return &OpCounter{n: initial}
}

func (c *OpCounter) Next() int32 {
	return atomic.AddInt32(&c.n, 1)
}

func (c *OpCounter) Reset(initial int32) {
	atomic.SwapInt32(&c.n, initial)
}

func (c *OpCounter) Current() int32 {
	return atomic.LoadInt32(&c.n)
}

func (c *OpCounter) Revert() {
	atomic.AddInt32(&c.n, -1)
}

type BufferedInterceptor struct {
	messages        *ReplayBuffer[*datalink.Message]
	confirmations   *ReplayBuffer[*datalink.Confirmation]
	baseInterceptor MessageInterceptor
	opCounter       *OpCounter
	handshake.DatabaseTransfer
}

func NewBufferedInterceptor(
	databaseTransfer handshake.DatabaseTransfer,
	interceptor MessageInterceptor,
) *BufferedInterceptor {
	return &BufferedInterceptor{
		baseInterceptor:  interceptor,
		DatabaseTransfer: databaseTransfer,
		opCounter:        NewOpCounter(0),
		messages:         NewReplayBuffer[*datalink.Message](MaxSize),
		confirmations:    NewReplayBuffer[*datalink.Confirmation](1000),
	}
}

func (o *BufferedInterceptor) OnMessage(message *datalink.Message) error {
	// todo verify message index if not the head
	message.MessageIndex = o.opCounter.Next()
	log.Println("Received message: ", message.MessageIndex)
	if err := o.messages.Add(message); err != nil {
		panic("illegal state")
	}
	return o.baseInterceptor.OnMessage(message)
}

func (o *BufferedInterceptor) OnConfirmation(confirmation *datalink.Confirmation) {
	log.Println("Received confirmation: ", confirmation.GetMessageIndex())
	if err := o.confirmations.Add(confirmation); err != nil {
		panic(err)
	}
	o.messages.ClearBefore(confirmation.GetMessageIndex()) // no need to keep old confirmed messages - every node has them
	o.baseInterceptor.OnConfirmation(confirmation)
}

func (o *BufferedInterceptor) GetMessagesAfter(i int32) []*datalink.Message {
	messages, err := o.messages.MessagesAfter(i)
	if err != nil {
		log.Fatalln(err)
	}
	return messages
}

func (o *BufferedInterceptor) GetConfirmationsAfter(i int32) []*datalink.Confirmation {
	confirmations, err := o.confirmations.MessagesAfter(i)
	if err != nil {
		log.Fatalln(err)
	}
	return confirmations
}

func (o *BufferedInterceptor) ProcessMessages(messages []*datalink.Message) {
	for _, msg := range messages {
		err := o.OnMessage(msg)
		if err != nil {
			log.Println("Failed to process message: ", err)
		}
	}
	err := o.messages.Add(messages...)
	if err != nil {
		log.Fatalln("Illegal state: ", err)
	}
}

func (o *BufferedInterceptor) ProcessConfirmations(confirmations []*datalink.Confirmation) {
	for _, conf := range confirmations {
		o.OnConfirmation(conf)
	}
	err := o.confirmations.Add(confirmations...)
	if err != nil {
		log.Fatalln("Illegal state: ", err)
	}
}

func (o *BufferedInterceptor) LastMessageIndex() int32 {
	i, err := o.messages.LastMessageIndex()
	if err != nil {
		return -1
	}
	return i
}

func (o *BufferedInterceptor) LastConfirmationIndex() int32 {
	i, err := o.confirmations.LastMessageIndex()
	if err != nil {
		return -1
	}
	return i
}

func (o *BufferedInterceptor) GetSnapshot() *datalink.DatabaseSnapshot {
	snapshot := o.DatabaseTransfer.GetSnapshot()
	snapshot.OpCount = o.opCounter.Current()
	return snapshot
}

func (o *BufferedInterceptor) SetFromSnapshot(snapshot *datalink.DatabaseSnapshot) {
	o.opCounter.Reset(snapshot.GetOpCount())
	o.DatabaseTransfer.SetFromSnapshot(snapshot)
}
