package handshake

import (
	"errors"
	"log"
	"seminarska/proto/datalink"

	"google.golang.org/grpc"
)

type serverStream = grpc.BidiStreamingServer[
	datalink.ClientHandshakeMsg,
	datalink.ServerHandshakeMsg,
]

type serverHandshake struct {
	data        ServerData
	stream      serverStream
	clientHello *datalink.ClientHello
}

func Server(stream serverStream, data ServerData) error {
	handshake := &serverHandshake{
		data:   data,
		stream: stream,
	}
	return run(handshake)
}

func (s *serverHandshake) receiveHello() error {
	received, err := s.stream.Recv()
	if err != nil {
		return err
	}
	clientHello, ok := received.Payload.(*datalink.ClientHandshakeMsg_Hello)
	if !ok {
		return errors.New("invalid handshake message: expected client hello")
	}
	s.clientHello = clientHello.Hello
	log.Println("Client: last confirmation index:", s.clientHello.GetLastConfIndex())
	return nil
}

func (s *serverHandshake) sendHello() error {
	return s.stream.Send(s.helloMsg())
}

func (s *serverHandshake) receiveMissingData() error {
	received, err := s.stream.Recv()
	if err != nil {
		return err
	}

	switch r := received.Payload.(type) {
	case *datalink.ClientHandshakeMsg_Sync:
		log.Println("Received missing messages")
		s.data.ProcessMessages(r.Sync.GetMessages())
	case *datalink.ClientHandshakeMsg_Db:
		log.Println("Received full DB snapshot")
		s.data.SetFromSnapshot(r.Db)
	default:
		return errors.New("invalid handshake message: expected client sync or db snapshot")
	}
	return nil
}

func (s *serverHandshake) sendMissingData() error {
	if s.data.LastMessageIndex() == -1 {
		return nil
	}
	hello := s.clientHello
	if hello == nil {
		panic("illegal data")
	}
	last := hello.GetLastConfIndex()
	return s.stream.Send(s.syncMsg(last))
}

func (s *serverHandshake) helloMsg() *datalink.ServerHandshakeMsg {
	lastMsg := s.data.LastMessageIndex()
	return &datalink.ServerHandshakeMsg{
		Payload: &datalink.ServerHandshakeMsg_Hello{
			Hello: &datalink.ServerHelo{
				LastMsgIndex:    lastMsg,
				RequestTransfer: lastMsg == -1,
			},
		},
	}
}

func (s *serverHandshake) syncMsg(after int32) *datalink.ServerHandshakeMsg {
	return &datalink.ServerHandshakeMsg{
		Payload: &datalink.ServerHandshakeMsg_Sync{
			Sync: &datalink.ServerSync{
				Confirmations: s.data.GetConfirmationsAfter(after),
			},
		},
	}
}
