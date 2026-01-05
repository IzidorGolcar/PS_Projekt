package keys

import (
	"testing"

	"seminarska/internal/data/storage/entities"
)

type testEntity struct {
	entities.BaseEntity
	Name string
}

func TestIndex_AddAndDuplicate(t *testing.T) {
	i := NewIndex("Name")
	e1 := &testEntity{}
	e1.SetId(1)
	e1.Name = "a"
	if err := i.Add(e1); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// duplicate id
	e2 := &testEntity{}
	e2.SetId(1)
	e2.Name = "b"
	if err := i.Add(e2); err == nil {
		t.Fatalf("expected duplicate id error")
	}
	// duplicate constraint (different id same name)
	e3 := &testEntity{}
	e3.SetId(2)
	e3.Name = "a"
	if err := i.Add(e3); err == nil {
		t.Fatalf("expected constraint error")
	}
}

func TestIndex_RemoveAndReplace(t *testing.T) {
	i := NewIndex("Name")
	e1 := &testEntity{}
	e1.SetId(10)
	e1.Name = "x"
	if err := i.Add(e1); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	i.Remove(e1)
	// now adding another with same name should succeed
	e2 := &testEntity{}
	e2.SetId(11)
	e2.Name = "x"
	if err := i.Add(e2); err != nil {
		t.Fatalf("unexpected err after remove: %v", err)
	}
	// test replace with same id but different name
	e2.Name = "y"
	if err := i.Replace(e2, e2); err != nil {
		t.Fatalf("replace should succeed: %v", err)
	}
}
