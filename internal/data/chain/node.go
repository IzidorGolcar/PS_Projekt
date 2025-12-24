package chain

import (
	"context"
	"log"
	"seminarska/internal/data/chain/handshake"
	"seminarska/proto/controllink"
	"seminarska/proto/datalink"
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

type Node struct {
	ctx         context.Context
	producer    MessageProducer
	chainClient *Client
	chainServer *Server
	done        chan struct{}
	state       *NodeDFA
	interceptor *BufferedInterceptor
}

func NewNode(
	ctx context.Context,
	messageProducer MessageProducer,
	messageInterceptor MessageInterceptor,
	transfer handshake.DatabaseTransfer,
	listenerAddress string,
) *Node {
	dfa := NewNodeDFA()
	interceptor := NewBufferedInterceptor(transfer, messageInterceptor)
	n := &Node{
		ctx:         ctx,
		producer:    messageProducer,
		done:        make(chan struct{}),
		state:       dfa,
		chainClient: NewClient(ctx, dfa, interceptor, 1000),
		chainServer: NewServer(ctx, dfa, listenerAddress, interceptor, 1000),
		interceptor: interceptor,
	}
	go n.run()
	return n
}

func (n *Node) run() {
	defer close(n.done)
	var (
		stateCtx context.Context
		cancel   context.CancelFunc
	)

	// FIXME when a mid node disconnect a message in its predecessors outbound channel will be sent twice

	for {
		select {
		case state := <-n.state.States():
			log.Println("Switching to state:", state)
			if cancel != nil {
				cancel()
			}
			stateCtx, cancel = context.WithCancel(n.ctx)
			switch state.Role {
			case Reader:
				go n.runAsHead(stateCtx)
			case Relay:
				go n.runAsMid(stateCtx)
			case Confirmer:
				go n.runAsTail(stateCtx)
			case ReaderConfirmer:
				go n.runAsSingleNode(stateCtx)
			}
		case <-n.ctx.Done():
			log.Println("Shutting down node")
			if cancel != nil {
				cancel()
			}
			return
		}
	}
}

func (n *Node) runAsHead(ctx context.Context) {
	for {
		select {
		case msg := <-n.producer.Messages():
			_ = n.interceptor.OnMessage(msg)
			n.chainClient.Outbound() <- msg
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
			_ = n.interceptor.OnMessage(msg)
			n.chainClient.Outbound() <- msg
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
			err := n.interceptor.OnMessage(msg)
			if err != nil {
				errConf := &datalink.Confirmation{
					MessageIndex: msg.GetMessageIndex(),
					RequestId:    msg.GetRequestId(),
					Ok:           false, Error: err.Error(),
				}
				n.interceptor.OnConfirmation(errConf)
				n.chainServer.Outbound() <- errConf
			} else {
				conf := &datalink.Confirmation{
					MessageIndex: msg.GetMessageIndex(),
					RequestId:    msg.GetRequestId(),
					Ok:           true,
				}
				n.interceptor.OnConfirmation(conf)
				n.chainServer.Outbound() <- conf
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
			err := n.interceptor.OnMessage(msg)
			if err != nil {
				log.Println("Failed to process message: ", err)
				n.interceptor.OnConfirmation(&datalink.Confirmation{
					MessageIndex: msg.GetMessageIndex(),
					Ok:           false, Error: err.Error(),
					RequestId: msg.GetRequestId(),
				})
			} else {
				n.interceptor.OnConfirmation(&datalink.Confirmation{
					MessageIndex: msg.GetMessageIndex(), Ok: true,
					RequestId: msg.GetRequestId(),
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

func (n *Node) SetRole(role controllink.NodeRole) error {
	switch role {
	case controllink.NodeRole_Relay:
		return n.state.Emit(RoleRelay)
	case controllink.NodeRole_MessageConfirmer:
		return n.state.Emit(RoleConfirmer)
	case controllink.NodeRole_MessageReader:
		return n.state.Emit(RoleReader)
	case controllink.NodeRole_MessageReaderConfirmer:
		return n.state.Emit(RoleReaderConfirmer)
	default:
		panic("illegal state")
	}
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
