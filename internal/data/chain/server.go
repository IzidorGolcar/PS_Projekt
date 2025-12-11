package chain

import (
	"context"
	"errors"
	"seminarska/internal/common/rpc"
	"seminarska/internal/data/storage"
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"

	"google.golang.org/grpc"
)

type listener struct {
	db *storage.Database
	datalink.UnimplementedDataLinkServer
}

func (l *listener) Register(grpcServer *grpc.Server) {
	datalink.RegisterDataLinkServer(grpcServer, l)
}

func (l *listener) WriteData(
	ctx context.Context,
	req *datalink.Record,
) (*datalink.DataWriteResponse, error) {
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
		//request = NewWriteRequest(p.Like)
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

	return &datalink.DataWriteResponse{}, nil
}

func (l *listener) IsRecordSynced(
	ctx context.Context,
	req *datalink.Record,
) (*datalink.DataSyncResponse, error) {

	var result any
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
		return nil, err
	}
	return &datalink.DataSyncResponse{Synced: result != nil}, nil
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
