package main

import (
	"log"
	"net"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	razpravljalnica.RegisterMessageBoardServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type server struct {
	razpravljalnica.UnimplementedMessageBoardServer
}

//func (s server) ListTopics(context.Context, *emptypb.Empty) (*razpravljalnica.ListTopicsResponse, error) {
//
//}
