package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"seminarska/internal/common/rpc"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/protobuf/types/known/emptypb"
)

// *************************************
// 				MOCK CLIENT
// *************************************

var nodes = []string{":5972", ":5982", ":5992"}

func main() {
	fmt.Println("Mock MessageBoard client")
	fmt.Println("Commands:")
	fmt.Println("  sub <node>")
	fmt.Println("  topic <node> <name>")
	fmt.Println("  topics <node>")
	fmt.Println("  msg <node> <topicId> <text>")
	fmt.Println("  exit")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			return
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		args := strings.SplitN(line, " ", 4)
		cmd := args[0]

		switch cmd {
		case "exit":
			return

		case "sub":
			requireArgs(args, 2)
			node := parseNode(args[1])
			subscribe(node)

		case "topic":
			requireArgs(args, 3)
			node := parseNode(args[1])
			createTopic(client(node), args[2])

		case "topics":
			requireArgs(args, 2)
			node := parseNode(args[1])
			readTopics(client(node))

		case "msg":
			requireArgs(args, 4)
			node := parseNode(args[1])
			topicID := parseInt64(args[2])
			postMessage(client(node), topicID, args[3])

		default:
			fmt.Println("unknown command")
		}
	}
}

func client(node int) razpravljalnica.MessageBoardClient {
	ctx := context.Background()
	cc := rpc.NewClient(ctx, nodes[node])
	return razpravljalnica.NewMessageBoardClient(cc)
}

func subscribe(node int) {
	ctx := context.Background()
	service := client(node)

	stream, err := service.SubscribeTopic(ctx, &razpravljalnica.SubscribeTopicRequest{
		UserId:  1,
		TopicId: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9},
	})
	if err != nil {
		log.Println("subscribe:", err)
		return
	}

	fmt.Println("Subscribed on node", node+1)

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Println("subscription ended:", err)
				return
			}
			fmt.Println("EVENT:", msg)
		}
	}()
}

func createTopic(client razpravljalnica.MessageBoardClient, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	topic, err := client.CreateTopic(ctx, &razpravljalnica.CreateTopicRequest{
		Name: name,
	})
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(topic)
}

func readTopics(client razpravljalnica.MessageBoardClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topics, err := client.ListTopics(ctx, &emptypb.Empty{})
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(topics)
}

func postMessage(
	client razpravljalnica.MessageBoardClient,
	topic int64,
	text string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	msg, err := client.PostMessage(ctx, &razpravljalnica.PostMessageRequest{
		UserId:  1,
		TopicId: topic,
		Text:    text,
	})
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(msg)
}

func requireArgs(args []string, n int) {
	if len(args) < n {
		panic("not enough arguments")
	}
}

func parseNode(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > len(nodes) {
		panic("invalid node")
	}
	return n - 1
}

func parseInt64(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic("invalid number")
	}
	return v
}
