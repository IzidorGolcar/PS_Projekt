package chain

import (
	"context"
	"errors"
	"log"
	"net"
	"seminarska/internal/common/rpc"
	"seminarska/internal/common/stream"
	"seminarska/internal/data/chain/handshake"
	"seminarska/proto/datalink"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type Server struct {
	l         *listener
	rpcServer *rpc.Server
}

func NewServer(
	ctx context.Context,
	state *nodeDFA,
	addr string,
	data handshake.ServerData,
	buffer int,
) *Server {
	l := newListener(state, data, buffer)
	return &Server{
		l:         l,
		rpcServer: rpc.NewServer(ctx, l, addr),
	}
}

func (s *Server) Outbound() chan<- *datalink.Confirmation {
	return s.l.outbound
}

func (s *Server) Inbound() <-chan *datalink.Message {
	return s.l.inbound
}

func (s *Server) Done() <-chan struct{} {
	return s.rpcServer.Done()
}

type session struct {
	addr          net.Addr
	cancel        context.CancelCauseFunc
	ctx           context.Context
	handshakeDone chan struct{}
}

type listener struct {
	datalink.UnimplementedDataLinkServer
	outbound chan *datalink.Confirmation
	inbound  chan *datalink.Message
	state    *nodeDFA
	data     handshake.ServerData

	mx             sync.Mutex
	currentSession *session
}

func newListener(
	state *nodeDFA,
	data handshake.ServerData,
	buffer int,
) *listener {
	return &listener{
		outbound: make(chan *datalink.Confirmation, buffer),
		inbound:  make(chan *datalink.Message, buffer),
		state:    state,
		data:     data,
	}
}

func (l *listener) Register(grpcServer *grpc.Server) {
	datalink.RegisterDataLinkServer(grpcServer, l)
}

func (l *listener) Handshake(s datalink.DataLink_HandshakeServer) error {
	p, ok := peer.FromContext(s.Context())
	if !ok {
		return errors.New("no peer info")
	}

	newSess := &session{
		addr:          p.Addr,
		handshakeDone: make(chan struct{}),
	}
	newSess.ctx, newSess.cancel = context.WithCancelCause(s.Context())

	l.mx.Lock()
	if l.currentSession != nil {
		log.Printf("Replacing connection: %s -> %s\n", l.currentSession.addr.String(), p.Addr.String())
		l.currentSession.cancel(errors.New("connection replaced by new client"))
	}
	l.currentSession = newSess
	l.mx.Unlock()

	err := handshake.Server(s, l.data)
	if err != nil {
		return err
	}

	close(newSess.handshakeDone)

	<-newSess.ctx.Done()
	return context.Cause(newSess.ctx)
}

func (l *listener) Replicate(s datalink.DataLink_ReplicateServer) error {
	p, ok := peer.FromContext(s.Context())
	if !ok {
		return errors.New("no peer info")
	}

	l.mx.Lock()
	sess := l.currentSession

	if sess == nil || sess.addr.String() != p.Addr.String() {
		l.mx.Unlock()
		return errors.New("handshake required before replication")
	}
	l.mx.Unlock()

	select {
	case <-sess.handshakeDone:
	case <-s.Context().Done():
		return s.Context().Err()
	case <-sess.ctx.Done():
		return context.Cause(sess.ctx)
	}

	l.state.emit(predecessorConnect)
	defer l.state.emit(predecessorDisconnect)

	supervisor := stream.NewSupervisor(l.outbound, l.inbound)

	return supervisor.Run(sess.ctx, s)
}
