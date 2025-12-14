package chain

import (
	"context"
	"errors"
	"seminarska/internal/common/rpc"
	"seminarska/internal/data/storage"
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type listener struct {
	db *storage.AppDatabase
	datalink.UnimplementedDataLinkServer
}

func (l *listener) Register(grpcServer *grpc.Server) {
	datalink.RegisterDataLinkServer(grpcServer, l)
}

func (l *listener) Write(_ context.Context, req *datalink.Record) (*emptypb.Empty, error) {
	record, err := entities.DatalinkToEntity(req)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	err = l.db.Insert(record)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (l *listener) Delete(_ context.Context, req *datalink.Record) (*emptypb.Empty, error) {
	record, err := entities.DatalinkToEntity(req)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	switch record.(type) {
	case *entities.User:
		err = l.db.DeleteUser(record.Id())
	case *entities.Message:
		err = l.db.DeleteMessage(record.Id())
	default:
		err = errors.New("invalid record type")
	}

	return &emptypb.Empty{}, err
}
func (l *listener) Update(_ context.Context, req *datalink.Record) (*emptypb.Empty, error) {
	record, err := entities.DatalinkToEntity(req)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	msg, ok := record.(*entities.Message)
	if !ok {
		return &emptypb.Empty{}, errors.New("invalid record type")
	}
	err = l.db.UpdateMessage(msg.Id(), func(current *entities.Message) (*entities.Message, error) {
		return msg, nil
	})
	return &emptypb.Empty{}, err
}

func (l *listener) Compare(
	ctx context.Context,
	req *datalink.Record,
) (*datalink.Comparison, error) {
	record, err := entities.DatalinkToEntity(req)
	if err != nil {
		return &datalink.Comparison{Equal: false}, err
	}
	var localRecord entities.Entity
	switch record.(type) {
	case *entities.User:
		localRecord, err = l.db.GetUser(record.Id())
	case *entities.Topic:
		localRecord, err = l.db.GetTopic(record.Id())
	case *entities.Message:
		localRecord, err = l.db.GetMessage(record.Id())
	case *entities.Like:
		localRecord, err = l.db.GetLike(record.Id())
	}

	if err != nil {
		return &datalink.Comparison{Equal: false}, err
	}
	equal := localRecord == record
	return &datalink.Comparison{Equal: equal}, nil
}

type Server struct {
	rpcServer *rpc.Server
}

func NewServer(ctx context.Context, address string) *Server {
	l := &listener{}
	return &Server{
		rpcServer: rpc.NewServer(ctx, l, address),
	}
}

func (s *Server) Done() <-chan struct{} {
	return s.rpcServer.Done()
}
