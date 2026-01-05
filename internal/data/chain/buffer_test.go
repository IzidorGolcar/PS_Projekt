package chain

import (
	"testing"
)

type idxMsg struct {
	idx int32
}

func (m *idxMsg) GetMessageIndex() int32 { return m.idx }

func TestReplayBuffer_AddAndOrder(t *testing.T) {
	b := NewReplayBuffer[*idxMsg](10)
	if err := b.Add(&idxMsg{idx: 1}, &idxMsg{idx: 2}); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	// out of order
	if err := b.Add(&idxMsg{idx: 2}); err == nil {
		t.Fatalf("expected ErrIndexOutOfOrder")
	}
	// last index
	li, err := b.LastMessageIndex()
	if err != nil {
		t.Fatalf("last index err: %v", err)
	}
	if li != 2 {
		t.Fatalf("expected last 2 got %d", li)
	}
}

func TestReplayBuffer_MessagesAfter(t *testing.T) {
	b := NewReplayBuffer[*idxMsg](10)
	_ = b.Add(&idxMsg{idx: 1}, &idxMsg{idx: 3}, &idxMsg{idx: 6})
	// exact match
	msgs, err := b.MessagesAfter(1)
	if err != nil {
		t.Fatalf("expected nil err for exact match: %v", err)
	}
	if len(msgs) != 2 || msgs[0].GetMessageIndex() != 3 {
		t.Fatalf("unexpected msgs: %v", msgs)
	}
	// greater than index (incomplete result)
	msgs2, err := b.MessagesAfter(4)
	if err == nil || err != ErrIncompleteResult {
		t.Fatalf("expected incomplete result, got %v %v", msgs2, err)
	}
	// no buffered message
	b2 := NewReplayBuffer[*idxMsg](10)
	_, err = b2.MessagesAfter(1)
	if err == nil || err != ErrNoBufferedMessages {
		t.Fatalf("expected no buffered messages: %v", err)
	}
}

func TestReplayBuffer_ClearBeforeAndSize(t *testing.T) {
	b := NewReplayBuffer[*idxMsg](2)
	_ = b.Add(&idxMsg{idx: 1}, &idxMsg{idx: 2}, &idxMsg{idx: 3})
	// size should trim to last 2
	li, _ := b.LastMessageIndex()
	if li != 3 {
		t.Fatalf("expected last 3 got %d", li)
	}
	b.ClearBefore(3)
	_, err := b.MessagesAfter(2)
	if !(err == ErrIncompleteResult || err == nil) {
		t.Fatalf("unexpected err: %v", err)
	}
	// ensure buffer starts at index >=3
	li2, err := b.LastMessageIndex()
	if err != nil || li2 < 3 {
		t.Fatalf("unexpected last after clear: %d %v", li2, err)
	}
}
