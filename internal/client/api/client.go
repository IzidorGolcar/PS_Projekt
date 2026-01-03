package api

import (
	"context"
	"seminarska/internal/common/rpc"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Client struct {
	control razpravljalnica.ControlPlaneClient
	ctx     context.Context
	userId  int
}

func (c *Client) UserId() int {
	return c.userId
}

func NewClient(ctx context.Context, controlAddress string) *Client {
	controlRpc := rpc.NewClient(ctx, controlAddress)
	control := razpravljalnica.NewControlPlaneClient(controlRpc)
	return &Client{
		ctx:     ctx,
		control: control,
	}
}

func (c *Client) headAddr() (string, error) {
	state, err := c.control.GetClusterState(c.ctx, &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return state.Head.GetAddress(), nil
}

func (c *Client) tailAddr() (string, error) {
	state, err := c.control.GetClusterState(c.ctx, &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return state.Tail.GetAddress(), nil
}

func (c *Client) subAddr() (string, string, error) {
	request := &razpravljalnica.SubscriptionNodeRequest{
		UserId:  int64(c.userId),
		TopicId: nil,
	}
	node, err := c.control.GetSubcscriptionNode(c.ctx, request)
	if err != nil {
		return "", "", err
	}
	return node.GetNode().GetAddress(), node.GetSubscribeToken(), nil
}

func (c *Client) getClient(addr string) razpravljalnica.MessageBoardClient {
	rpcClient := rpc.NewClient(c.ctx, addr)
	return razpravljalnica.NewMessageBoardClient(rpcClient)
}

func (c *Client) SignUp(username string) error {
	addr, err := c.headAddr()
	if err != nil {
		return err
	}
	req := &razpravljalnica.CreateUserRequest{
		Name: username,
	}
	res, err := c.getClient(addr).CreateUser(c.ctx, req)
	if err != nil {
		return err
	}
	c.userId = int(res.GetId())
	return nil
}

func (c *Client) Login(username string) error {
	addr, err := c.tailAddr()
	if err != nil {
		return err
	}
	name := username
	req := &razpravljalnica.GetUserRequest{
		Username: &name,
	}
	user, err := c.getClient(addr).GetUser(c.ctx, req)
	if err != nil {
		return err
	}
	c.userId = int(user.GetId())
	return nil
}

func (c *Client) ListTopics() ([]*razpravljalnica.Topic, error) {
	addr, err := c.tailAddr()
	if err != nil {
		return nil, err
	}
	topics, err := c.getClient(addr).ListTopics(c.ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	return topics.GetTopics(), nil
}

func (c *Client) CreateTopic(name string) error {
	addr, err := c.headAddr()
	if err != nil {
		return err
	}
	req := &razpravljalnica.CreateTopicRequest{
		Name: name,
	}
	_, err = c.getClient(addr).CreateTopic(c.ctx, req)
	return err
}

func (c *Client) GetUsername(userId int) (string, error) {
	addr, err := c.tailAddr()
	if err != nil {
		return "", err
	}
	id := int64(userId)
	req := &razpravljalnica.GetUserRequest{
		UserId: &id,
	}
	user, err := c.getClient(addr).GetUser(c.ctx, req)
	if err != nil {
		return "", err
	}
	return user.GetName(), nil
}

func (c *Client) GetMessages(topicId int) ([]*razpravljalnica.Message, error) {
	addr, err := c.tailAddr()
	if err != nil {
		return nil, err
	}
	req := &razpravljalnica.GetMessagesRequest{
		TopicId:       int64(topicId),
		FromMessageId: 0,
		Limit:         0,
	}
	messages, err := c.getClient(addr).GetMessages(c.ctx, req)
	if err != nil {
		return nil, err
	}
	return messages.GetMessages(), nil
}

func (c *Client) PostMessage(topicId int, text string) error {
	addr, err := c.headAddr()
	if err != nil {
		return err
	}
	req := &razpravljalnica.PostMessageRequest{
		TopicId: int64(topicId),
		UserId:  int64(c.userId),
		Text:    text,
	}
	_, err = c.getClient(addr).PostMessage(c.ctx, req)
	return err
}

func (c *Client) Subscribe(ctx context.Context, topicId int) (<-chan *razpravljalnica.Message, error) {
	addr, token, err := c.subAddr()
	if err != nil {
		return nil, err
	}
	req := &razpravljalnica.SubscribeTopicRequest{
		TopicId:        []int64{int64(topicId)},
		UserId:         int64(c.userId),
		FromMessageId:  0,
		SubscribeToken: token,
	}
	res, err := c.getClient(addr).SubscribeTopic(ctx, req)
	if err != nil {
		return nil, err
	}
	ch := make(chan *razpravljalnica.Message, 100)
	go func() {
		defer close(ch)
		for {
			msg, err := res.Recv()
			if err != nil {
				return
			}
			ch <- msg.GetMessage()
		}
	}()
	return ch, nil
}
