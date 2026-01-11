package handshake

import (
	"errors"
	"log"
	"seminarska/proto/datalink"

	"google.golang.org/grpc"
)

type clientStream = grpc.BidiStreamingClient[
	datalink.ClientHandshakeMsg,
	datalink.ServerHandshakeMsg,
]

type clientHandshake struct {
	stream      clientStream
	data        ClientData
	serverHello *datalink.ServerHelo
}

func Client(stream clientStream, data ClientData) error {
	handshake := &clientHandshake{
		stream: stream,
		data:   data,
	}
	return run(handshake)
}

func (c *clientHandshake) sendHello() error {
	return c.stream.Send(c.helloMsg())
}

func (c *clientHandshake) receiveHello() error {
	received, err := c.stream.Recv()
	if err != nil {
		return err
	}
	serverHello, ok := received.Payload.(*datalink.ServerHandshakeMsg_Hello)
	if !ok {
		return errors.New("invalid handshake message: expected server hello")
	}
	c.serverHello = serverHello.Hello
	return nil
}

func (c *clientHandshake) sendMissingData() error {
	hello := c.serverHello
	if hello.GetRequestTransfer() {
		log.Println("Sending full DB snapshot")
		return c.stream.Send(c.snapshotMsg())
	}
	log.Println("successor: last message index:", c.serverHello.GetLastMsgIndex())
	return c.stream.Send(c.syncMsg(hello.GetLastMsgIndex()))
}

func (c *clientHandshake) receiveMissingData() error {
	if c.serverHello.GetRequestTransfer() {
		return nil
	}
	received, err := c.stream.Recv()
	if err != nil {
		return err
	}
	confirmations, ok := received.Payload.(*datalink.ServerHandshakeMsg_Sync)
	if !ok {
		return errors.New("invalid handshake message: expected server sync")
	}
	log.Println("Received missing confirmations")
	c.data.ProcessConfirmations(confirmations.Sync.GetConfirmations())
	return nil
}

func (c *clientHandshake) helloMsg() *datalink.ClientHandshakeMsg {
	return &datalink.ClientHandshakeMsg{
		Payload: &datalink.ClientHandshakeMsg_Hello{
			Hello: &datalink.ClientHello{
				LastConfIndex: c.data.LastConfirmationIndex(),
			},
		},
	}
}

func (c *clientHandshake) syncMsg(after int32) *datalink.ClientHandshakeMsg {
	return &datalink.ClientHandshakeMsg{
		Payload: &datalink.ClientHandshakeMsg_Sync{
			Sync: &datalink.ClientSync{
				Messages: c.data.GetMessagesAfter(after),
			},
		},
	}
}

func (c *clientHandshake) snapshotMsg() *datalink.ClientHandshakeMsg {
	return &datalink.ClientHandshakeMsg{
		Payload: &datalink.ClientHandshakeMsg_Db{
			Db: c.data.GetSnapshot(),
		},
	}
}
