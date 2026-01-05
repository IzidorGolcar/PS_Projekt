package entities

import (
	"testing"
	"time"

	"seminarska/proto/datalink"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestDatalinkToEntity_UserRoundtrip(t *testing.T) {
	dl := &datalink.Message{Payload: &datalink.Message_User{User: &razpravljalnica.User{Id: 7, Name: "alice"}}}
	e, err := DatalinkToEntity(dl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	u, ok := e.(*User)
	if !ok {
		t.Fatalf("expected *User, got %T", e)
	}
	if u.Id() != 7 || u.Name != "alice" {
		t.Fatalf("unexpected user fields: %+v", u)
	}

	back := EntityToDatalink(u)
	if back.GetUser() == nil || back.GetUser().Id != 7 || back.GetUser().Name != "alice" {
		t.Fatalf("roundtrip failed: %v", back)
	}
}

func TestDatalinkToEntity_MessageRoundtrip(t *testing.T) {
	tm := time.Now().Truncate(time.Second)
	protoMsg := &razpravljalnica.Message{Id: 11, TopicId: 2, UserId: 3, Text: "hi", CreatedAt: timestamppb.New(tm)}
	dl := &datalink.Message{Payload: &datalink.Message_Message{Message: protoMsg}}
	e, err := DatalinkToEntity(dl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := e.(*Message)
	if !ok {
		t.Fatalf("expected *Message, got %T", e)
	}
	if m.Id() != 11 || m.TopicId != 2 || m.UserId != 3 || m.Text != "hi" || !m.CreatedAt.Equal(tm) {
		t.Fatalf("unexpected message fields: %+v", m)
	}

	back := EntityToDatalink(m)
	bm := back.GetMessage()
	if bm == nil || bm.Id != 11 || bm.TopicId != 2 || bm.UserId != 3 || bm.Text != "hi" {
		t.Fatalf("roundtrip failed: %v", back)
	}
}

func TestDatalinkToEntity_TopicAndLike(t *testing.T) {
	dlt := &datalink.Message{Payload: &datalink.Message_Topic{Topic: &razpravljalnica.Topic{Id: 5, Name: "go"}}}
	e, err := DatalinkToEntity(dlt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tp, ok := e.(*Topic)
	if !ok || tp.Id() != 5 || tp.Name != "go" {
		t.Fatalf("unexpected topic: %+v", e)
	}
	if got := EntityToDatalink(tp).GetTopic(); got == nil || got.Id != 5 || got.Name != "go" {
		t.Fatalf("topic roundtrip failed: %v", got)
	}

	dll := &datalink.Message{Payload: &datalink.Message_Like{Like: &razpravljalnica.Like{Id: 9, UserId: 4, MessageId: 8}}}
	e2, err := DatalinkToEntity(dll)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lk, ok := e2.(*Like)
	if !ok || lk.Id() != 9 || lk.UserId != 4 || lk.MessageId != 8 {
		t.Fatalf("unexpected like: %+v", lk)
	}
	if got := EntityToDatalink(lk).GetLike(); got == nil || got.UserId != 4 || got.MessageId != 8 {
		t.Fatalf("like roundtrip failed: %v", got)
	}
}
