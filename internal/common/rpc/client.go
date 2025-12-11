package rpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	conn   *grpc.ClientConn
}

func (c *Client) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	return c.conn.Invoke(ctx, method, args, reply, opts...)
}

func (c *Client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return c.conn.NewStream(ctx, desc, method, opts...)
}

func NewClient(ctx context.Context, addr string) *Client {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	clientCtx, cancel := context.WithCancel(ctx)
	c := &Client{
		ctx:    clientCtx,
		cancel: cancel,
		conn:   conn,
		done:   make(chan struct{}),
	}
	go c.run()
	return c
}

func (c *Client) run() {
	<-c.ctx.Done()
	err := c.conn.Close()
	if err != nil {
		log.Printf("failed to close connection: %v", err)
	}
	close(c.done)
}

func (c *Client) Stop() {
	c.cancel()
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}
