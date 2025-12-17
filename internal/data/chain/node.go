package chain

import (
	"context"
	"errors"
	"log"
	"seminarska/proto/datalink"
)

type NodeState int

const (
	StateHead NodeState = iota
	StateMiddle
	StateTail
	StateSingleNode
)

// MessageProducer produces messages at the head of the chain
type MessageProducer interface {
	Messages() <-chan *datalink.Message
}

// MessageInterceptor intercepts messages and confirmations at each node of the chain
type MessageInterceptor interface {
	OnMessage(message *datalink.Message) error
	OnConfirmation(confirmation *datalink.Confirmation)
}

type UniversalChainNode interface {
	MessageProducer
	MessageInterceptor
}

type OpCounter struct {
	n int
}

func NewOpCounter(initial int) *OpCounter {
	return &OpCounter{n: initial}
}

func (c *OpCounter) Next() int32 {
	current := c.n
	c.n++
	return int32(current)
}

type Node struct {
	ctx         context.Context
	interceptor MessageInterceptor
	producer    MessageProducer
	chainClient *Client
	chainServer *Server
	done        chan struct{}
	state       chan NodeState
	counter     *OpCounter
}

func NewNode(
	ctx context.Context,
	chain UniversalChainNode,
	listenerAddress string,
) *Node {
	n := &Node{
		ctx:         ctx,
		interceptor: chain,
		producer:    chain,
		done:        make(chan struct{}),
		state:       make(chan NodeState, 1),
		chainClient: NewClient(ctx, 1000),
		chainServer: NewServer(ctx, listenerAddress, 1000),
		counter:     NewOpCounter(0),
	}
	go n.run()
	return n
}

func (n *Node) SetState(state NodeState) {
	n.state <- state
}

func (n *Node) run() {
	defer close(n.done)
	var (
		stateCtx context.Context
		cancel   context.CancelFunc
	)
	for {
		select {
		case state := <-n.state:
			if cancel != nil {
				cancel()
			}
			stateCtx, cancel = context.WithCancel(n.ctx)
			switch state {
			case StateHead:
				go n.runAsHead(stateCtx)
			case StateMiddle:
				go n.runAsMid(stateCtx)
			case StateTail:
				go n.runAsTail(stateCtx)
			case StateSingleNode:
				go n.runAsSingleNode(stateCtx)
			}
		case <-n.ctx.Done():
			if cancel != nil {
				cancel()
				return
			}
		}
	}
}

func (n *Node) runAsHead(ctx context.Context) {
	for {
		select {
		case msg := <-n.producer.Messages():
			msg.MessageIndex = n.counter.Next()
			err := n.interceptor.OnMessage(msg)
			if err != nil {
				log.Println("Failed to process message: ", err)
			} else {
				n.chainClient.Outbound() <- msg
			}
		case conf := <-n.chainClient.Inbound():
			n.interceptor.OnConfirmation(conf)
		case <-ctx.Done():
			return
		}
	}
}

func (n *Node) runAsMid(ctx context.Context) {
	for {
		select {
		case msg := <-n.chainServer.Inbound():
			if msg.GetMessageIndex() != n.counter.Next() {
				err := errors.New("message not synced")
				errConf := &datalink.Confirmation{MessageIndex: msg.GetMessageIndex(), Ok: false, Error: err.Error()}
				n.interceptor.OnConfirmation(errConf)
			}
			err := n.interceptor.OnMessage(msg)
			if err != nil {
				log.Println("Failed to process message: ", err)
				errConf := &datalink.Confirmation{MessageIndex: msg.GetMessageIndex(), Ok: false, Error: err.Error()}
				n.chainServer.Outbound() <- errConf
			} else {
				n.chainClient.Outbound() <- msg
			}
		case conf := <-n.chainClient.Inbound():
			n.interceptor.OnConfirmation(conf)
			n.chainServer.Outbound() <- conf
		case <-ctx.Done():
			return
		}
	}
}

func (n *Node) runAsTail(ctx context.Context) {
	for {
		select {
		case msg := <-n.chainServer.Inbound():
			if msg.GetMessageIndex() != n.counter.Next() {
				err := errors.New("message not synced")
				errConf := &datalink.Confirmation{MessageIndex: msg.GetMessageIndex(), Ok: false, Error: err.Error()}
				n.interceptor.OnConfirmation(errConf)
			}
			err := n.interceptor.OnMessage(msg)
			if err != nil {
				log.Println("Failed to process message: ", err)
				errConf := &datalink.Confirmation{MessageIndex: msg.GetMessageIndex(), Ok: false, Error: err.Error()}
				n.chainServer.Outbound() <- errConf
			}
		case <-ctx.Done():
			return
		}
	}
}

func (n *Node) runAsSingleNode(ctx context.Context) {
	for {
		select {
		case msg := <-n.producer.Messages():
			msg.MessageIndex = n.counter.Next()
			err := n.interceptor.OnMessage(msg)
			if err != nil {
				log.Println("Failed to process message: ", err)
			} else {
				n.interceptor.OnConfirmation(&datalink.Confirmation{
					MessageIndex: msg.GetMessageIndex(), Ok: true,
				})
			}
		case <-ctx.Done():
			return
		}
	}
}

func (n *Node) SetNextNode(addr string) error {
	return n.chainClient.SetNextNode(addr)
}

func (n *Node) Done() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		<-n.done
		<-n.chainClient.Done()
		<-n.chainServer.Done()
		close(done)
	}()
	return done
}
