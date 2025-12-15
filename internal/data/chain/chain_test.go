package chain

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"seminarska/internal/common/rpc"
	"seminarska/proto/datalink"
)

// ==== Test helpers ====

func withTimeout(t *testing.T, d time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	if d == 0 {
		d = 5 * time.Second
	}
	return context.WithTimeout(context.Background(), d)
}

func freeTCPAddr(t *testing.T) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get free addr: %v", err)
	}
	addr := lis.Addr().String()
	_ = lis.Close()
	return addr
}

// fakeStream is a minimal in-memory bidirectional stream implementing Stream[O,I].
type fakeStream[O any, I any] struct {
	// Inputs to the stream (from supervisor's transmit)
	sent chan O
	// Outputs from the stream (to supervisor's receive)
	recv chan I
	// If non-nil, Send returns this error after first send attempt
	sendErr atomic.Value // holds error or nil
	// If non-nil, Recv returns this error immediately when no value is available
	recvErr atomic.Value // holds error or nil
}

func newFakeStream[O any, I any]() *fakeStream[O, I] {
	return &fakeStream[O, I]{
		sent: make(chan O, 16),
		recv: make(chan I, 16),
	}
}

func (f *fakeStream[O, I]) Send(o O) error {
	if perr, _ := f.sendErr.Load().(error); perr != nil {
		return perr
	}
	f.sent <- o
	return nil
}

func (f *fakeStream[O, I]) Recv() (I, error) {
	var zero I
	if perr, _ := f.recvErr.Load().(error); perr != nil {
		return zero, perr
	}
	i, ok := <-f.recv
	if !ok {
		return zero, errors.New("stream closed")
	}
	return i, nil
}

// ==== StreamSupervisor tests ====

func TestStreamSupervisor_SendAndReceive(t *testing.T) {
	ctx, cancel := withTimeout(t, 3*time.Second)
	defer cancel()

	outbound := make(chan int, 8)
	inbound := make(chan string, 8)
	ss := NewStreamSupervisor(outbound, inbound)
	fs := newFakeStream[int, string]()

	// Preload one incoming message and then close receive side after a bit
	fs.recv <- "ok"

	// Run supervisor
	done := make(chan struct{})
	go func() {
		_ = ss.Run(ctx, fs)
		close(done)
	}()

	// Send a value; it should arrive on fakeStream.sent
	outbound <- 42

	select {
	case got := <-fs.sent:
		if got != 42 {
			t.Fatalf("expected 42 sent to stream, got %v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for stream send")
	}

	// The preloaded recv value should arrive on inbound
	select {
	case got := <-inbound:
		if got != "ok" {
			t.Fatalf("expected 'ok' on inbound, got %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for inbound receive")
	}

	// Cancel the context and expect Run to finish
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("supervisor did not stop after context cancel")
	}
}

func TestStreamSupervisor_SendErrorCancelsAndStoresDropped(t *testing.T) {
	ctx, cancel := withTimeout(t, 3*time.Second)
	defer cancel()

	outbound := make(chan int, 1)
	inbound := make(chan int, 1)
	ss := NewStreamSupervisor(outbound, inbound)
	fs := newFakeStream[int, int]()

	// Configure send to fail
	errSend := errors.New("boom")
	fs.sendErr.Store(errSend)

	done := make(chan struct{})
	go func() {
		_ = ss.Run(ctx, fs)
		close(done)
	}()

	outbound <- 7

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("supervisor did not stop after send error")
	}
	if ss.DroppedMessage() == nil || *ss.DroppedMessage() != 7 {
		t.Fatalf("expected dropped message = 7, got %v", ss.DroppedMessage())
	}
}

func TestStreamSupervisor_RecvErrorCancels(t *testing.T) {
	ctx, cancel := withTimeout(t, 3*time.Second)
	defer cancel()

	outbound := make(chan int, 1)
	inbound := make(chan int, 1)
	ss := NewStreamSupervisor(outbound, inbound)
	fs := newFakeStream[int, int]()

	// Configure receive to fail
	errRecv := errors.New("recv-fail")
	fs.recvErr.Store(errRecv)

	done := make(chan struct{})
	go func() {
		_ = ss.Run(ctx, fs)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("supervisor did not stop after recv error")
	}
}

// ==== Server tests (with real gRPC) ====

func TestServer_ReplicateDataFlow(t *testing.T) {
	addr := freeTCPAddr(t)
	ctx, cancel := withTimeout(t, 5*time.Second)
	defer cancel()

	srv := NewServer(ctx, addr, 8)

	// Create a real gRPC client against same addr
	client := rpc.NewClient(ctx, addr)
	link := datalink.NewDataLinkClient(client)

	stream, err := link.Replicate(ctx)
	if err != nil {
		t.Fatalf("Replicate failed: %v", err)
	}

	// 1) Send message from client to server and verify it arrives on server.Inbound()
	msg := &datalink.Message{MessageId: 1}
	if err := stream.Send(msg); err != nil {
		t.Fatalf("stream.Send failed: %v", err)
	}

	select {
	case got := <-srv.Inbound():
		if got.GetMessageId() != 1 {
			t.Fatalf("expected message id=1, got %v", got.GetMessageId())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message on server inbound")
	}

	// 2) Send confirmation from server to client and read via stream.Recv()
	conf := &datalink.Confirmation{MessageId: 1}
	srv.Outbound() <- conf

	select {
	case got := <-func() <-chan *datalink.Confirmation {
		ch := make(chan *datalink.Confirmation, 1)
		go func() {
			g, _ := stream.Recv()
			ch <- g
		}()
		return ch
	}():
		if got.GetMessageId() != 1 {
			t.Fatalf("expected confirmation messageId=1, got: %v", got.GetMessageId())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for confirmation on client stream")
	}
}

// ==== Client tests (with real Server) ====

func TestClient_ConnectSendReceive(t *testing.T) {
	// Start a server
	addr := freeTCPAddr(t)
	sctx, scancel := withTimeout(t, 10*time.Second)
	defer scancel()
	srv := NewServer(sctx, addr, 8)

	// Start client
	cctx, ccancel := withTimeout(t, 10*time.Second)
	defer ccancel()
	cl := NewClient(cctx, 8)

	if err := cl.SetNextNode(addr); err != nil {
		t.Fatalf("SetNextNode failed: %v", err)
	}

	// Send a message from client to server
	wantMsg := &datalink.Message{MessageId: 7}
	cl.Outbound() <- wantMsg

	select {
	case got := <-srv.Inbound():
		if got.GetMessageId() != wantMsg.GetMessageId() {
			t.Fatalf("server got wrong message id: %v", got.GetMessageId())
		}
	case <-time.After(4 * time.Second):
		t.Fatal("timeout waiting for message at server inbound")
	}

	// Send a confirmation from server to client
	srv.Outbound() <- &datalink.Confirmation{MessageId: wantMsg.GetMessageId()}

	select {
	case got := <-cl.Inbound():
		if got.GetMessageId() != wantMsg.GetMessageId() {
			t.Fatalf("client got wrong confirmation id: %v", got.GetMessageId())
		}
	case <-time.After(4 * time.Second):
		t.Fatal("timeout waiting for confirmation at client inbound")
	}
}

func TestClient_AddressChangeReconnects(t *testing.T) {
	// Start first server
	addr1 := freeTCPAddr(t)
	ctx1, cancel1 := withTimeout(t, 10*time.Second)
	defer cancel1()
	srv1 := NewServer(ctx1, addr1, 8)

	// Start second server
	addr2 := freeTCPAddr(t)
	ctx2, cancel2 := withTimeout(t, 10*time.Second)
	defer cancel2()
	srv2 := NewServer(ctx2, addr2, 8)

	t.Log("srv1:", addr1, "srv2:", addr2)

	// Client
	cctx, ccancel := withTimeout(t, 15*time.Second)
	defer ccancel()
	cl := NewClient(cctx, 8)

	if err := cl.SetNextNode(addr1); err != nil {
		t.Fatalf("SetNextNode addr1 failed: %v", err)
	}

	// Ensure messages go to server 1
	cl.Outbound() <- &datalink.Message{MessageId: 1}
	select {
	case got := <-srv1.Inbound():
		if got.GetMessageId() != 1 {
			t.Fatalf("expected id=1 on srv1, got %v", got.GetMessageId())
		}
	case <-time.After(4 * time.Second):
		t.Fatal("timeout waiting for message at server1 inbound")
	}

	// Switch to server 2
	if err := cl.SetNextNode(addr2); err != nil {
		t.Fatalf("SetNextNode addr2 failed: %v", err)
	}

	time.Sleep(time.Microsecond)

	// Subsequent messages should arrive to server 2
	cl.Outbound() <- &datalink.Message{MessageId: 2}
	select {
	case got := <-srv2.Inbound():
		if got.GetMessageId() != 2 {
			t.Fatalf("expected id=2 on srv2, got %v", got.GetMessageId())
		}
	case <-time.After(6 * time.Second):
		t.Fatal("timeout waiting for message at server2 inbound")
	}

	// Try to see that server1 does not receive the new message within short period
	select {
	case got := <-srv1.Inbound():
		t.Fatalf("unexpected message on server1 after address change: id=%v", got.GetMessageId())
	case <-time.After(500 * time.Millisecond):
		// ok, no message
	}
}
