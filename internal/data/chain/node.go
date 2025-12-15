package chain

import (
	"context"
	"seminarska/proto/datalink"
)

const (
	StateHead = iota
	StateMiddle
	StateTail
)

type MessageInterceptor interface {
	OnMessage(message *datalink.Message) error
	OnConfirmation(confirmation *datalink.Confirmation)
}

type Node struct {
	ctx         context.Context
	interceptor MessageInterceptor
	chainClient *Client
	chainServer *Server
	state       int
}

func NewNode(
	ctx context.Context,
	interceptor MessageInterceptor,
	listenerAddress string,
) *Node {
	n := &Node{
		ctx:         ctx,
		interceptor: interceptor,
		chainClient: NewClient(ctx, 1000),
		chainServer: NewServer(ctx, listenerAddress, 1000),
	}
	go n.run()
	return n
}

func (n *Node) run() {
	go n.runUpstream()
	go n.runDownstream()
}

func (n *Node) runUpstream() {
	for msg := range n.chainServer.Inbound() {
		err := n.interceptor.OnMessage(msg)
		if err != nil {
			errStr := err.Error()
			n.chainServer.Outbound() <- &datalink.Confirmation{MessageId: msg.MessageId, Ok: false, Error: &errStr}
		} else {
			n.chainClient.Outbound() <- msg
		}
	}
}

func (n *Node) runDownstream() {
	for conf := range n.chainClient.Inbound() {
		n.interceptor.OnConfirmation(conf)
		n.chainServer.Outbound() <- conf
	}
}

func (n *Node) SetNextNode(addr string) error {
	return n.chainClient.SetNextNode(addr)
}

func (n *Node) Done() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		<-n.chainClient.Done()
		<-n.chainServer.Done()
		close(done)
	}()
	return done
}
