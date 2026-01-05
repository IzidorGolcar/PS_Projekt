package chain

import (
	"context"
	"errors"
	"log"
	"seminarska/internal/common/rpc"
	"seminarska/internal/data/chain/handshake"
	"seminarska/internal/data/chain/stream"
	"seminarska/proto/datalink"
	"sync"
	"time"
)

type Client struct {
	ctx      context.Context
	state    *NodeDFA
	addr     chan string
	requests chan *datalink.Message
	replies  chan *datalink.Confirmation
	data     handshake.ClientData
	done     chan struct{}
}

func NewClient(
	ctx context.Context,
	state *NodeDFA,
	data handshake.ClientData,
	buffer int,
) *Client {
	c := &Client{
		ctx:      ctx,
		state:    state,
		data:     data,
		addr:     make(chan string),
		requests: make(chan *datalink.Message, buffer),
		replies:  make(chan *datalink.Confirmation, buffer),
		done:     make(chan struct{}),
	}
	go c.run()
	return c
}

var (
	errAddressChange = errors.New("address changed")
)

func (c *Client) run() {
	defer close(c.done)
	defer close(c.addr)

	var (
		connectionCtx context.Context
		cancel        context.CancelCauseFunc
	)

	connectionMutex := &sync.Mutex{}

	for {
		select {
		case addr := <-c.addr:
			if cancel != nil {
				cancel(errAddressChange)
				// todo wait for disconnect
			}
			if addr != "" {
				connectionCtx, cancel = context.WithCancelCause(c.ctx)
				go func() {
					connectionMutex.Lock()
					defer connectionMutex.Unlock()
					c.superviseConnection(addr, connectionCtx)
				}()
			}
		case <-c.ctx.Done():
			if cancel != nil {
				cancel(nil)
			}
			return
		}
	}
}

func (c *Client) superviseConnection(addr string, ctx context.Context) {
	if err := c.state.Emit(SuccessorConnect); err != nil {
		panic(err)
	}
	defer func() {
		if err := c.state.Emit(SuccessorDisconnect); err != nil {
			panic(err)
		}
	}()

	for {
		log.Println("datalink connecting to ", addr)
		rpcClient := rpc.NewClient(ctx, addr)
		link := datalink.NewDataLinkClient(rpcClient)

		if err := c.doHandshake(link); err != nil {
			log.Println("handshake failed: ", err, " retrying in 5 seconds")
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		if err := c.superviseStream(link, ctx); err != nil {
			if errors.Is(err, errAddressChange) ||
				errors.Is(err, context.Canceled) {
				return
			}
			log.Println("connection failed: ", err, " retrying in 5 seconds")
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}
	}

}

func (c *Client) doHandshake(link datalink.DataLinkClient) error {
	handshakeStream, err := link.Handshake(c.ctx)
	if err != nil {
		return err
	}
	return handshake.Client(handshakeStream, c.data)
}

func (c *Client) superviseStream(link datalink.DataLinkClient, ctx context.Context) error {
	s, err := link.Replicate(ctx)
	if err != nil {
		return err
	}
	supervisor := stream.NewSupervisor(c.requests, c.replies)
	defer func() {
		if supervisor.DroppedMessage() != nil {
			panic("dropped message")
		}
	}()
	return supervisor.Run(ctx, s)
}

func (c *Client) SetNextNode(addr string) error {
	select {
	case c.addr <- addr:
		return nil
	case <-c.done:
		return errors.New("chain is closed")
	}
}

func (c *Client) Outbound() chan<- *datalink.Message {
	return c.requests
}

func (c *Client) Inbound() <-chan *datalink.Confirmation {
	return c.replies
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}
