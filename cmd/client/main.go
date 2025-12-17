package main

import (
	"context"
	"fmt"
	"log"
	"seminarska/internal/common/rpc"
	"seminarska/proto/razpravljalnica"
)

func main() {
	ctx := context.Background()
	client := rpc.NewClient(ctx, ":5972")
	service := razpravljalnica.NewMessageBoardClient(client)

	topic, err := createTopic(ctx, service)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(topic)
}

func createTopic(ctx context.Context, client razpravljalnica.MessageBoardClient) (*razpravljalnica.Topic, error) {
	return client.CreateTopic(ctx, &razpravljalnica.CreateTopicRequest{Name: "test"})
}
