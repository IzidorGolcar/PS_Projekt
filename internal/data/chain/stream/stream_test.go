package stream

import (
	"context"
	"errors"
	"testing"
	"time"
)

// minimal fake stream that echoes sends to a channel and provides recvs from a channel
type fakeStream[O any, I any] struct {
	sendErr error
	recvErr error
	sendCh  chan O
	recvCh  chan I
}

func (f *fakeStream[O, I]) Send(req O) error {
	if f.sendErr != nil {
		return f.sendErr
	}
	f.sendCh <- req
	return nil
}
func (f *fakeStream[O, I]) Recv() (I, error) {
	if f.recvErr != nil {
		var z I
		return z, f.recvErr
	}
	v := <-f.recvCh
	return v, nil
}

func TestSupervisor_SendReceive_Success(t *testing.T) {
	out := make(chan int, 1)
	in := make(chan string, 1)
	s := NewSupervisor[int, string](out, in)
	fs := &fakeStream[int, string]{sendCh: make(chan int, 1), recvCh: make(chan string, 1)}
	ctx, cancel := context.WithCancelCause(context.Background())
	// run supervisor in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Run(ctx, fs)
	}()

	// put a reply to be received
	fs.recvCh <- "ok"
	// send outbound message
	out <- 42

	// give some time for transmit/recv to process
	time.Sleep(10 * time.Millisecond)
	// cancel context so Run returns
	cancel(nil)

	err := <-errCh
	if err == nil {
		t.Fatalf("expected non-nil error (canceled), got nil")
	}
}

func TestSupervisor_SendFailure(t *testing.T) {
	out := make(chan int, 1)
	in := make(chan string, 1)
	s := NewSupervisor[int, string](out, in)
	fs := &fakeStream[int, string]{sendErr: errors.New("sendfail"), sendCh: make(chan int, 1), recvCh: make(chan string, 1)}
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	// run supervisor in goroutine so we can wait for error
	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(ctx, fs) }()

	// enqueue a message to trigger send
	out <- 7

	// wait for run to abort due to send failure
	err := <-errCh
	if err == nil {
		t.Fatalf("expected error from run due to send failure")
	}
	if s.DroppedMessage() == nil {
		t.Fatalf("expected dropped message to be set")
	}
}

func TestSupervisor_ReceiveFailure(t *testing.T) {
	out := make(chan int, 1)
	in := make(chan string, 1)
	s := NewSupervisor[int, string](out, in)
	fs := &fakeStream[int, string]{recvErr: errors.New("recvfail"), sendCh: make(chan int, 1), recvCh: make(chan string, 1)}
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	// run supervisor in goroutine
	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(ctx, fs) }()

	// enqueue an outbound message so transmit starts; receive will error
	out <- 8

	// expect run to return with error
	err := <-errCh
	if err == nil {
		t.Fatalf("expected error from run due to recv failure")
	}
}
