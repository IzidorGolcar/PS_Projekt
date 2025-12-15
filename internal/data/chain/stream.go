package chain

import (
	"context"
	"errors"
)

type Stream[Req any, Res any] interface {
	Send(req Req) error
	Recv() (Res, error)
}

type StreamSupervisor[O any, I any] struct {
	outbound       chan O
	inbound        chan I
	droppedMessage *O
}

func NewStreamSupervisor[O any, I any](
	outbound chan O,
	inbound chan I,
) *StreamSupervisor[O, I] {
	return &StreamSupervisor[O, I]{outbound, inbound, nil}
}

func (c *StreamSupervisor[O, I]) DroppedMessage() *O {
	return c.droppedMessage
}

func (c *StreamSupervisor[O, I]) Run(ctx context.Context, stream Stream[O, I]) error {
	streamCtx, cancel := context.WithCancelCause(ctx)
	go c.transmit(stream, streamCtx, cancel)
	go c.receive(stream, streamCtx, cancel)
	<-streamCtx.Done()
	return streamCtx.Err()
}

func (c *StreamSupervisor[O, I]) transmit(
	stream Stream[O, I],
	ctx context.Context,
	cancel context.CancelCauseFunc,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.outbound:
			if err := stream.Send(msg); err != nil {
				c.droppedMessage = &msg
				cancel(errors.Join(errors.New("failed to send message"), err))
				return
			}
		}
	}
}

func (c *StreamSupervisor[O, I]) receive(
	stream Stream[O, I],
	ctx context.Context,
	cancel context.CancelCauseFunc,
) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			cancel(errors.Join(errors.New("failed to receive confirmation"), err))
			return
		}
		c.inbound <- msg
		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}
