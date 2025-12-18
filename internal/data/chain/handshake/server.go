package handshake

import (
	"errors"
	"seminarska/proto/datalink"

	"google.golang.org/grpc"
)

type serverStream grpc.BidiStreamingServer[datalink.ClientHandshakeMsg, datalink.ServerHandshakeMsg]

type Server struct {
	state NodeState
}

func NewServer(state NodeState) *Server {
	return &Server{state: state}
}

func (s *Server) Run(stream serverStream) error {
	received, err := stream.Recv()
	if err != nil {
		return err
	}
	clientHello, ok := received.Payload.(*datalink.ClientHandshakeMsg_Hello)
	if !ok {
		return errors.New("invalid handshake message: expected client hello")
	}
	go s.confSync(stream, clientHello)

	clientSync, ok := received.Payload.(*datalink.ClientHandshakeMsg_Sync)
	if !ok {
		return errors.New("invalid handshake message: expected client sync")
	}

	messages := clientSync.Sync.GetMessages()
	panic("todo")
}

func (s *Server) confSync(stream serverStream, msg *datalink.ClientHandshakeMsg_Hello) error {
	last := msg.Hello.GetLastConfIndex()
	return stream.Send(s.syncMsg(last))
}

func (s *Server) helloMsg() *datalink.ServerHandshakeMsg {
	return &datalink.ServerHandshakeMsg{
		Payload: &datalink.ServerHandshakeMsg_Hello{
			Hello: &datalink.ServerHelo{
				LastMsgIndex: s.state.LastMessageIndex(),
			},
		},
	}
}

func (s *Server) syncMsg(after int32) *datalink.ServerHandshakeMsg {
	return &datalink.ServerHandshakeMsg{
		Payload: &datalink.ServerHandshakeMsg_Sync{
			Sync: &datalink.ServerSync{
				Confirmations: s.state.GetConfirmationsAfter(after),
			},
		},
	}
}
