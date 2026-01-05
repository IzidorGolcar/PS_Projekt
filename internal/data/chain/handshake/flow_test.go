package handshake

import (
	"errors"
	"testing"
)

type fakeProvider struct {
	calls  []string
	failAt string
}

func (f *fakeProvider) sendHello() error {
	f.calls = append(f.calls, "sendHello")
	if f.failAt == "sendHello" {
		return errors.New("fail")
	}
	return nil
}
func (f *fakeProvider) receiveHello() error {
	f.calls = append(f.calls, "receiveHello")
	if f.failAt == "receiveHello" {
		return errors.New("fail")
	}
	return nil
}
func (f *fakeProvider) sendMissingData() error {
	f.calls = append(f.calls, "sendMissingData")
	if f.failAt == "sendMissingData" {
		return errors.New("fail")
	}
	return nil
}
func (f *fakeProvider) receiveMissingData() error {
	f.calls = append(f.calls, "receiveMissingData")
	if f.failAt == "receiveMissingData" {
		return errors.New("fail")
	}
	return nil
}

func TestRun_Success(t *testing.T) {
	p := &fakeProvider{}
	if err := run(p); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(p.calls) != 4 {
		t.Fatalf("expected 4 calls got %d", len(p.calls))
	}
}

func TestRun_Failure(t *testing.T) {
	p := &fakeProvider{failAt: "receiveHello"}
	if err := run(p); err == nil {
		t.Fatalf("expected error")
	}
	if p.calls[0] != "sendHello" || p.calls[1] != "receiveHello" {
		t.Fatalf("unexpected call order: %v", p.calls)
	}
}
