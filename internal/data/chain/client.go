package chain

import (
	"context"
	"errors"
	"log"
	"seminarska/internal/common/rpc"
	"seminarska/internal/common/stream"
	"seminarska/proto/datalink"
	"time"
)

type Client struct {
	ctx      context.Context
	state    *nodeDFA
	addr     chan string
	requests chan *datalink.Message
	replies  chan *datalink.Confirmation
	done     chan struct{}
}

func NewClient(
	ctx context.Context,
	state *nodeDFA,
	buffer int,
) *Client {
	c := &Client{
		ctx:      ctx,
		state:    state,
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

	for {
		select {
		case addr := <-c.addr:
			if cancel != nil {
				cancel(errAddressChange)
			}
			if addr != "" {
				connectionCtx, cancel = context.WithCancelCause(c.ctx)
				go c.superviseConnection(addr, connectionCtx)
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
	for {
		log.Println("datalink connecting to ", addr)
		rpcClient := rpc.NewClient(ctx, addr)
		link := datalink.NewDataLinkClient(rpcClient)

		// TODO: perform sync with next node, switch state

		err := c.superviseStream(link, ctx)
		if err != nil {
			if errors.Is(err, errAddressChange) ||
				errors.Is(err, context.Canceled) ||
				errors.Is(err, context.DeadlineExceeded) {
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

func (c *Client) superviseStream(link datalink.DataLinkClient, ctx context.Context) error {
	s, err := link.Replicate(ctx)
	if err != nil {
		return err
	}
	c.state.emit(successorConnect)
	defer c.state.emit(successorDisconnect)
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
