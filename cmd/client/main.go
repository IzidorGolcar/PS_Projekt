package main

import (
	"context"
	"fmt"
	"log"
	"seminarska/internal/common/rpc"
	"seminarska/proto/razpravljalnica"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	for {
		run()
	}
}

func run() {

	nodes := []string{":5972", ":5982", ":5992"}
	var head, tail int
	var name string
	fmt.Scanf("%d %d %s", &head, &tail, &name)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	headClient := rpc.NewClient(ctx, nodes[head-1])
	headService := razpravljalnica.NewMessageBoardClient(headClient)

	tailClient := rpc.NewClient(ctx, nodes[tail-1])
	tailService := razpravljalnica.NewMessageBoardClient(tailClient)

	createTopic(ctx, headService, name)
	readTopics(ctx, tailService)
}

func createTopic(ctx context.Context, client razpravljalnica.MessageBoardClient, name string) {
	topic, err := client.CreateTopic(ctx, &razpravljalnica.CreateTopicRequest{Name: name})
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(topic)
	}
}

func readTopics(ctx context.Context, client razpravljalnica.MessageBoardClient) {
	topics, err := client.ListTopics(ctx, &emptypb.Empty{})
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(topics)
	}
}
