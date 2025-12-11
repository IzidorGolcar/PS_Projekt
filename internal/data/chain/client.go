package chain

import (
	"context"
	"errors"
	"seminarska/internal/common/rpc"
	"seminarska/proto/datalink"
	"seminarska/proto/razpravljalnica"
)

type Client struct {
	ctx         context.Context
	rpcClient   *rpc.Client
	chainClient datalink.DataLinkClient
}

func NewClient(ctx context.Context, addr string) *Client {
	rpcClient := rpc.NewClient(ctx, addr)
	return &Client{
		ctx:         ctx,
		rpcClient:   rpcClient,
		chainClient: datalink.NewDataLinkClient(rpcClient),
	}
}

func (c *Client) Done() <-chan struct{} {
	return c.rpcClient.Done()
}

func getRecord(data any) (*datalink.Record, error) {
	var request *datalink.Record
	switch r := data.(type) {
	case *razpravljalnica.User:
		request = &datalink.Record{Payload: &datalink.Record_User{User: r}}
	case *razpravljalnica.Message:
		request = &datalink.Record{Payload: &datalink.Record_Message{Message: r}}
	case *razpravljalnica.Like:
		request = &datalink.Record{Payload: &datalink.Record_Like{Like: r}}
	case *razpravljalnica.Topic:
		request = &datalink.Record{Payload: &datalink.Record_Topic{Topic: r}}
	default:
		return nil, errors.New("invalid payload")
	}
	return request, nil
}

func (c *Client) WriteData(data any) error {
	record, err := getRecord(data)
	if err != nil {
		return err
	}
	_, err = c.chainClient.WriteData(c.ctx, record)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Sync(data any) (bool, error) {
	record, err := getRecord(data)
	if err != nil {
		return false, err
	}
	res, err := c.chainClient.IsRecordSynced(c.ctx, record)
	if err != nil {
		return false, err
	}
	return res.Synced, nil
}
