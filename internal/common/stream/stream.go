package stream

import (
	"context"
	"errors"
	"log"
)

type BidiStream[Req any, Res any] interface {
	Send(req Req) error
	Recv() (Res, error)
}

type Supervisor[O any, I any] struct {
	outbound       chan O
	inbound        chan I
	droppedMessage *O
}

func NewSupervisor[O any, I any](
	outbound chan O,
	inbound chan I,
) *Supervisor[O, I] {
	return &Supervisor[O, I]{outbound, inbound, nil}
}

func (c *Supervisor[O, I]) DroppedMessage() *O {
	return c.droppedMessage
}

func (c *Supervisor[O, I]) Run(ctx context.Context, stream BidiStream[O, I]) error {
	streamCtx, cancel := context.WithCancelCause(ctx)
	go c.transmit(stream, streamCtx, cancel)
	go c.receive(stream, streamCtx, cancel)
	<-streamCtx.Done()
	return streamCtx.Err()
}

func (c *Supervisor[O, I]) transmit(
	stream BidiStream[O, I],
	ctx context.Context,
	cancel context.CancelCauseFunc,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.outbound:
			log.Println("sending to stream: ", msg)
			if err := stream.Send(msg); err != nil {
				log.Println("failed to send to stream: ", err)
				c.droppedMessage = &msg
				cancel(errors.Join(errors.New("failed to send message"), err))
				return
			}
		}
	}
}

func (c *Supervisor[O, I]) receive(
	stream BidiStream[O, I],
	ctx context.Context,
	cancel context.CancelCauseFunc,
) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Println("failed to receive from stream: ", err)
			cancel(errors.Join(errors.New("failed to receive confirmation"), err))
			return
		}
		log.Println("received from stream: ", msg)
		c.inbound <- msg
		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}
