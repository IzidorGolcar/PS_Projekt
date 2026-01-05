package chain

import (
	"testing"

	"seminarska/proto/datalink"
)

type fakeTransfer struct {
	lastMsg int32
}

func (f *fakeTransfer) GetSnapshot() *datalink.DatabaseSnapshot    { return &datalink.DatabaseSnapshot{} }
func (f *fakeTransfer) SetFromSnapshot(*datalink.DatabaseSnapshot) {}

type nopInterceptor struct{}

func (n *nopInterceptor) OnMessage(*datalink.Message) error     { return nil }
func (n *nopInterceptor) OnConfirmation(*datalink.Confirmation) {}

func TestBufferedInterceptor_Basic(t *testing.T) {
	bi := NewBufferedInterceptor(&fakeTransfer{}, &nopInterceptor{})
	// messages without index should get assigned
	if err := bi.OnMessage(&datalink.Message{RequestId: "r1"}); err != nil {
		t.Fatalf("onmessage: %v", err)
	}
	if bi.LastMessageIndex() != 1 {
		t.Fatalf("expected last 1 got %d", bi.LastMessageIndex())
	}
	// confirmations
	if err := bi.confirmations.Add(&datalink.Confirmation{MessageIndex: 1, RequestId: "r1", Ok: true}); err != nil {
		t.Fatalf("add conf: %v", err)
	}
	if bi.LastConfirmationIndex() != 1 {
		t.Fatalf("expected last conf 1 got %d", bi.LastConfirmationIndex())
	}
}

func TestBufferedInterceptor_ProcessAndGetAfter(t *testing.T) {
	bi := NewBufferedInterceptor(&fakeTransfer{}, &nopInterceptor{})
	msgs := []*datalink.Message{{MessageIndex: 10}, {MessageIndex: 11}}
	bi.ProcessMessages(msgs)
	if got := bi.GetMessagesAfter(9); len(got) != 2 {
		t.Fatalf("expected 2 got %d", len(got))
	}
	confs := []*datalink.Confirmation{{MessageIndex: 20}, {MessageIndex: 21}}
	bi.ProcessConfirmations(confs)
	if got := bi.GetConfirmationsAfter(19); len(got) != 2 {
		t.Fatalf("expected 2 confs got %d", len(got))
	}
}

func TestBufferedInterceptor_SnapshotReset(t *testing.T) {
	bi := NewBufferedInterceptor(&fakeTransfer{}, &nopInterceptor{})
	bi.opCounter.Reset(5)
	snap := bi.GetSnapshot()
	if snap.GetOpCount() != 5 {
		t.Fatalf("expected snap opcount 5 got %d", snap.GetOpCount())
	}
	bi.SetFromSnapshot(&datalink.DatabaseSnapshot{OpCount: 2})
	if bi.opCounter.Current() != 2 {
		t.Fatalf("expected opcount 2 got %d", bi.opCounter.Current())
	}
}
