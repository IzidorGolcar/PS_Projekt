package chain

import (
	"context"
	"errors"
	"seminarska/internal/common/rpc"
	"seminarska/internal/data/storage"
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type listener struct {
	db *storage.AppDatabase
	datalink.UnimplementedDataLinkServer
}

func (l *listener) Register(grpcServer *grpc.Server) {
	datalink.RegisterDataLinkServer(grpcServer, l)
}

func (l *listener) Write(
	ctx context.Context,
	req *datalink.Record,
) (*emptypb.Empty, error) {
	var record entities.Entity
	switch p := req.Payload.(type) {
	case *datalink.Record_User:
		record = entities.NewUser(p.User.Name)
		record.SetId(p.User.Id)
	case *datalink.Record_Message:
		record = entities.NewMessage(
			p.Message.TopicId, p.Message.UserId,
			p.Message.Text, p.Message.CreatedAt.AsTime(),
		)
		record.SetId(p.Message.Id)
	case *datalink.Record_Like:

	case *datalink.Record_Topic:
		record = entities.NewTopic(p.Topic.Name)
		record.SetId(p.Topic.Id)
	default:
		return nil, errors.New("invalid payload")
	}

	err := l.db.Insert(record)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (l *listener) Delete(ctx context.Context, req *datalink.Record) (*emptypb.Empty, error) {
	var err error
	switch p := req.Payload.(type) {
	case *datalink.Record_User:
		err = l.db.DeleteUser(p.User.Id)
	default:
		return nil, errors.New("invalid payload")
	}

	return &emptypb.Empty{}, err
}
func (l *listener) Update(context.Context, *datalink.Record) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "method Update not implemented")
}

func (l *listener) Compare(
	ctx context.Context,
	req *datalink.Record,
) (*datalink.Comparison, error) {
	var result entities.Entity
	var err error
	switch p := req.Payload.(type) {
	case *datalink.Record_User:
		result, err = l.db.GetUser(p.User.Id)
	case *datalink.Record_Message:
		result, err = l.db.GetMessage(p.Message.Id)
	case *datalink.Record_Topic:
		result, err = l.db.GetTopic(p.Topic.Id)
	default:
		return nil, errors.New("invalid payload")
	}
	if err != nil {
		return &datalink.Comparison{Equal: false}, err
	}
	equal := result.ToDatalinkRecord().Payload == req.Payload
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
