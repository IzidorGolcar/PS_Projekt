package chain

import (
	"context"
	"seminarska/internal/common/rpc"
	"seminarska/proto/datalink"
)

type Client struct {
	ctx       context.Context
	rpcClient *rpc.Client
	datalink.DataLinkClient
}

func NewClient(ctx context.Context, addr string) *Client {
	rpcClient := rpc.NewClient(ctx, addr)
	return &Client{
		ctx:            ctx,
		rpcClient:      rpcClient,
		DataLinkClient: datalink.NewDataLinkClient(rpcClient),
	}
}

func (c *Client) Done() <-chan struct{} {
	return c.rpcClient.Done()
}
