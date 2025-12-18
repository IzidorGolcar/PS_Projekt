package handshake

import (
	"context"
	"errors"
	"seminarska/proto/datalink"

	"google.golang.org/grpc"
)

type clientStream grpc.BidiStreamingClient[datalink.ClientHandshakeMsg, datalink.ServerHandshakeMsg]

type Client struct {
	link  datalink.DataLinkClient
	state NodeState
}

func NewClient(
	link datalink.DataLinkClient,
	state NodeState,
) *Client {
	return &Client{
		link:  link,
		state: state,
	}
}

func (c *Client) Run(ctx context.Context) error {
	handshakeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := c.link.Handshake(handshakeCtx)
	if err != nil {
		return err
	}

	if err = stream.Send(c.helloMsg()); err != nil {
		return err
	}

	received, err := stream.Recv()
	if err != nil {
		return err
	}
	serverHello, ok := received.Payload.(*datalink.ServerHandshakeMsg_Hello)
	if !ok {
		return errors.New("invalid handshake message: expected server hello")
	}
	go c.msgSync(stream, serverHello)

	received, err = stream.Recv()
	if err != nil {
		return err
	}
	serverSync, ok := received.Payload.(*datalink.ServerHandshakeMsg_Sync)
	if !ok {
		return errors.New("invalid handshake message: expected server sync")
	}
	c.confSync(serverSync)

	return ctx.Err()
}

func (c *Client) msgSync(stream clientStream, msg *datalink.ServerHandshakeMsg_Hello) error {
	last := msg.Hello.GetLastMsgIndex()
	if last == -1 {
		return stream.Send(c.copyMsg())
	}
	return stream.Send(c.syncMsg(last))
}

func (c *Client) confSync(msg *datalink.ServerHandshakeMsg_Sync) error {
	confirmations := msg.Sync.GetConfirmations()
}

func (c *Client) helloMsg() *datalink.ClientHandshakeMsg {
	return &datalink.ClientHandshakeMsg{
		Payload: &datalink.ClientHandshakeMsg_Hello{
			Hello: &datalink.ClientHello{
				LastConfIndex: c.state.LastConfirmationIndex(),
			},
		},
	}
}

func (c *Client) syncMsg(after int32) *datalink.ClientHandshakeMsg {
	return &datalink.ClientHandshakeMsg{
		Payload: &datalink.ClientHandshakeMsg_Sync{
			Sync: &datalink.ClientSync{
				Messages: c.state.GetMessagesAfter(after),
			},
		},
	}
}

func (c *Client) copyMsg() *datalink.ClientHandshakeMsg {
	return &datalink.ClientHandshakeMsg{
		Payload: &datalink.ClientHandshakeMsg_Db{
			Db: &datalink.DatabaseCopy{},
		},
	}
}
