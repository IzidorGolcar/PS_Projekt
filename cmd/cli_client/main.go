package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"seminarska/internal/common/rpc"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/protobuf/types/known/emptypb"
)

// ************************************
// 				CLI CLIENT
// ************************************

var controlClient razpravljalnica.ControlPlaneClient
var currentUser *razpravljalnica.User

func main() {
	addr := flag.String("addr", ":7000", "Control service address")
	flag.Parse()

	fmt.Println("MessageBoard CLI client")
	fmt.Printf("Connecting to control service at %s...\n", *addr)

	ctx := context.Background()
	cc := rpc.NewClient(ctx, *addr)
	controlClient = razpravljalnica.NewControlPlaneClient(cc)

	printHelp()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		if currentUser != nil {
			fmt.Printf("[%s] > ", currentUser.Name)
		} else {
			fmt.Print("> ")
		}

		if !scanner.Scan() {
			return
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		args := strings.Fields(line)
		cmd := args[0]

		switch cmd {
		case "exit":
			return
		case "help":
			printHelp()
		case "user":
			if !requireArgs(args, 2) {
				continue
			}
			createUser(args[1])
		case "login":
			if !requireArgs(args, 2) {
				continue
			}
			login(args[1])
		case "topic":
			if !requireArgs(args, 2) {
				continue
			}
			createTopic(args[1])
		case "topics":
			listTopics()
		case "msg":
			if !requireArgs(args, 3) {
				continue
			}
			postMessage(parseInt64(args[1]), strings.Join(args[2:], " "))
		case "msgs":
			if !requireArgs(args, 2) {
				continue
			}
			topicID := parseInt64(args[1])
			fromID := int64(0)
			limit := int32(10)
			if len(args) > 2 {
				fromID = parseInt64(args[2])
			}
			if len(args) > 3 {
				limit = int32(parseInt64(args[3]))
			}
			getMessages(topicID, fromID, limit)
		case "update":
			if !requireArgs(args, 4) {
				continue
			}
			updateMessage(parseInt64(args[1]), parseInt64(args[2]), strings.Join(args[3:], " "))
		case "delete":
			if !requireArgs(args, 3) {
				continue
			}
			deleteMessage(parseInt64(args[1]), parseInt64(args[2]))
		case "like":
			if !requireArgs(args, 3) {
				continue
			}
			likeMessage(parseInt64(args[1]), parseInt64(args[2]))
		case "sub":
			if !requireArgs(args, 2) {
				continue
			}
			var topicIDs []int64
			for _, s := range args[1:] {
				topicIDs = append(topicIDs, parseInt64(s))
			}
			subscribe(topicIDs)
		default:
			fmt.Println("unknown command, type 'help' for usage")
		}
	}
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  user <name>                     - Create a new user")
	fmt.Println("  login <name>                    - Login as an existing user")
	fmt.Println("  topic <name>                    - Create a new topic")
	fmt.Println("  topics                          - List all topics")
	fmt.Println("  msg <topicId> <text>            - Post a message to a topic")
	fmt.Println("  msgs <topicId> [fromId] [limit] - Get messages from a topic")
	fmt.Println("  update <topicId> <msgId> <text> - Update a message")
	fmt.Println("  delete <topicId> <msgId>        - Delete a message")
	fmt.Println("  like <topicId> <msgId>          - Like a message")
	fmt.Println("  sub <topicId1> [topicId2]...    - Subscribe to topic(s)")
	fmt.Println("  help                            - Show this help")
	fmt.Println("  exit                            - Exit the client")
}

func getHeadClient() (razpravljalnica.MessageBoardClient, error) {
	state, err := controlClient.GetClusterState(context.Background(), &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get head address: %v", err)
	}
	cc := rpc.NewClient(context.Background(), state.Head.Address)
	return razpravljalnica.NewMessageBoardClient(cc), nil
}

func getTailClient() (razpravljalnica.MessageBoardClient, error) {
	state, err := controlClient.GetClusterState(context.Background(), &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get tail address: %v", err)
	}
	cc := rpc.NewClient(context.Background(), state.Tail.Address)
	return razpravljalnica.NewMessageBoardClient(cc), nil
}

func createUser(name string) {
	client, err := getHeadClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	user, err := client.CreateUser(context.Background(), &razpravljalnica.CreateUserRequest{Name: name})
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}
	currentUser = user
	fmt.Printf("User created: ID=%d, Name=%s\n", user.Id, user.Name)
}

func login(name string) {
	client, err := getTailClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	user, err := client.GetUser(context.Background(), &razpravljalnica.GetUserRequest{Username: &name})
	if err != nil {
		fmt.Printf("Error logging in: %v\n", err)
		return
	}
	currentUser = user
	fmt.Printf("Logged in as: ID=%d, Name=%s\n", user.Id, user.Name)
}

func createTopic(name string) {
	client, err := getHeadClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	topic, err := client.CreateTopic(context.Background(), &razpravljalnica.CreateTopicRequest{Name: name})
	if err != nil {
		fmt.Printf("Error creating topic: %v\n", err)
		return
	}
	fmt.Printf("Topic created: ID=%d, Name=%s\n", topic.Id, topic.Name)
}

func listTopics() {
	client, err := getTailClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.ListTopics(context.Background(), &emptypb.Empty{})
	if err != nil {
		fmt.Printf("Error listing topics: %v\n", err)
		return
	}
	fmt.Println("Topics:")
	for _, t := range res.Topics {
		fmt.Printf("  [%d] %s\n", t.Id, t.Name)
	}
}

func postMessage(topicID int64, text string) {
	if currentUser == nil {
		fmt.Println("You must be logged in to post messages")
		return
	}
	client, err := getHeadClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	msg, err := client.PostMessage(context.Background(), &razpravljalnica.PostMessageRequest{
		TopicId: topicID,
		UserId:  currentUser.Id,
		Text:    text,
	})
	if err != nil {
		fmt.Printf("Error posting message: %v\n", err)
		return
	}
	fmt.Printf("Message posted: ID=%d\n", msg.Id)
}

func getMessages(topicID int64, fromID int64, limit int32) {
	client, err := getTailClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.GetMessages(context.Background(), &razpravljalnica.GetMessagesRequest{
		TopicId:       topicID,
		FromMessageId: fromID,
		Limit:         limit,
	})
	if err != nil {
		fmt.Printf("Error getting messages: %v\n", err)
		return
	}
	fmt.Printf("Messages for topic %d:\n", topicID)
	for _, m := range res.Messages {
		fmt.Printf("  [%d] User %d: %s (likes: %d)\n", m.Id, m.UserId, m.Text, m.Likes)
	}
}

func updateMessage(topicID int64, msgID int64, text string) {
	if currentUser == nil {
		fmt.Println("You must be logged in to update messages")
		return
	}
	client, err := getHeadClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	msg, err := client.UpdateMessage(context.Background(), &razpravljalnica.UpdateMessageRequest{
		TopicId:   topicID,
		UserId:    currentUser.Id,
		MessageId: msgID,
		Text:      text,
	})
	if err != nil {
		fmt.Printf("Error updating message: %v\n", err)
		return
	}
	fmt.Printf("Message updated: ID=%d\n", msg.Id)
}

func deleteMessage(topicID int64, msgID int64) {
	if currentUser == nil {
		fmt.Println("You must be logged in to delete messages")
		return
	}
	client, err := getHeadClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = client.DeleteMessage(context.Background(), &razpravljalnica.DeleteMessageRequest{
		TopicId:   topicID,
		UserId:    currentUser.Id,
		MessageId: msgID,
	})
	if err != nil {
		fmt.Printf("Error deleting message: %v\n", err)
		return
	}
	fmt.Println("Message deleted")
}

func likeMessage(topicID int64, msgID int64) {
	if currentUser == nil {
		fmt.Println("You must be logged in to like messages")
		return
	}
	client, err := getHeadClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	msg, err := client.LikeMessage(context.Background(), &razpravljalnica.LikeMessageRequest{
		TopicId:   topicID,
		MessageId: msgID,
		UserId:    currentUser.Id,
	})
	if err != nil {
		fmt.Printf("Error liking message: %v\n", err)
		return
	}
	fmt.Printf("Message liked. New likes: %d\n", msg.Likes)
}

func subscribe(topicIDs []int64) {
	if currentUser == nil {
		fmt.Println("You must be logged in to subscribe")
		return
	}
	res, err := controlClient.GetSubcscriptionNode(context.Background(), &razpravljalnica.SubscriptionNodeRequest{
		UserId:  currentUser.Id,
		TopicId: topicIDs,
	})
	if err != nil {
		fmt.Printf("Error getting subscription node: %v\n", err)
		return
	}

	cc := rpc.NewClient(context.Background(), res.Node.Address)
	service := razpravljalnica.NewMessageBoardClient(cc)

	stream, err := service.SubscribeTopic(context.Background(), &razpravljalnica.SubscribeTopicRequest{
		UserId:         currentUser.Id,
		TopicId:        topicIDs,
		SubscribeToken: res.SubscribeToken,
	})
	if err != nil {
		fmt.Printf("Error subscribing: %v\n", err)
		return
	}

	fmt.Printf("Subscribed on node %s\n", res.Node.Address)

	go func() {
		for {
			event, err := stream.Recv()
			if err != nil {
				fmt.Printf("\nSubscription ended: %v\n", err)
				return
			}
			fmt.Printf("\n[EVENT] %v: %v\n> ", event.Op, event.Message)
		}
	}()
}

func requireArgs(args []string, n int) bool {
	if len(args) < n {
		fmt.Printf("Not enough arguments. Expected at least %d, got %d\n", n-1, len(args)-1)
		return false
	}
	return true
}

func parseInt64(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		fmt.Printf("Invalid number: %s\n", s)
		return 0
	}
	return v
}
