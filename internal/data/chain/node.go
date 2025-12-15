package chain

import (
	"context"
	"fmt"
	"log"
	"seminarska/internal/data/storage/entities"
	"time"
)

type Node struct {
	ctx         context.Context
	next        *connectedNode
	tail        *connectedNode
	chainServer *Server
}

func NewNode(ctx context.Context, listenerAddress string) *Node {
	return &Node{
		ctx:         ctx,
		chainServer: NewServer(ctx, listenerAddress),
	}
}

func datalinkContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*5)
}

func (n *Node) Forward(entity entities.Entity) error {
	if n.next == nil {
		return nil
	}
	ctx, cancel := datalinkContext()
	defer cancel()
	_, err := n.next.Write(ctx, entities.EntityToDatalink(entity))
	return err
}

func (n *Node) Delete(entity entities.Entity) error {
	if n.next == nil {
		return nil
	}
	ctx, cancel := datalinkContext()
	defer cancel()
	_, err := n.next.Delete(ctx, entities.EntityToDatalink(entity))
	return err
}

func (n *Node) Update(entity entities.Entity) error {
	if n.next == nil {
		return nil
	}
	ctx, cancel := datalinkContext()
	defer cancel()
	_, err := n.next.Update(ctx, entities.EntityToDatalink(entity))
	return err
}

func (n *Node) Compare(entity entities.Entity) (bool, error) {
	if n.tail == nil {
		return true, nil
	}
	ctx, cancel := datalinkContext()
	defer cancel()
	cmp, err := n.tail.Compare(ctx, entities.EntityToDatalink(entity))
	if err != nil {
		return false, err
	}
	return cmp.Equal, nil
}

func (n *Node) SetNextNode(address string) {
	if n.next != nil {
		err := n.next.disconnect(time.Second)
		if err != nil {
			log.Println(fmt.Errorf("failed to disconnect from next node: %w", err))
		}
	}
	log.Printf("Connecting to next node [%s]\n", address)
	n.next = newConnectedNode(n.ctx, address)
}

func (n *Node) SetTail(address string) {
	if n.tail != nil {
		err := n.tail.disconnect(time.Second)
		if err != nil {
			log.Println(fmt.Errorf("failed to disconnect from tail: %w", err))
		}
	}
	log.Printf("Connecting to tail node [%s]\n", address)
	n.tail = newConnectedNode(n.ctx, address)
}

func (n *Node) Done() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		<-n.chainServer.Done()
		if n.next != nil {
			<-n.next.Done()
		}
		if n.tail != nil {
			<-n.tail.Done()
		}
		close(done)
	}()
	return done
}
