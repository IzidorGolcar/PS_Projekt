package db

import (
	"testing"
)

// simple entity for tests
type e struct {
	id   int64
	Name string
}

func (x *e) Id() int64      { return x.id }
func (x *e) SetId(id int64) { x.id = id }

func TestRelation_InsertGet_Count(t *testing.T) {
	r := NewRelation[*e]()
	val := &e{Name: "a"}
	val.SetId(1)
	ins, err := r.Insert(val)
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	// confirm insertion so Get will see it
	if err := ins.Confirm(); err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	got, err := r.Get(1)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Id() != 1 || got.Name != "a" {
		t.Fatalf("unexpected get: %+v", got)
	}
	if c := r.Count(); c != 1 {
		t.Fatalf("expected count 1 got %d", c)
	}
}

func TestRelation_DeleteAndConfirm(t *testing.T) {
	r := NewRelation[*e]()
	val := &e{Name: "del"}
	val.SetId(2)
	ins, err := r.Insert(val)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	// confirm insertion
	if err := ins.Confirm(); err != nil {
		t.Fatalf("confirm insert: %v", err)
	}
	// delete
	delRec, err := r.Delete(2)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	// DeletedValue returns the deleted value
	dv := delRec.DeletedValue()
	if dv.Id() != 2 {
		t.Fatalf("deleted value id mismatch: %v", dv.Id())
	}
	if err := delRec.Confirm(); err != nil {
		t.Fatalf("confirm delete: %v", err)
	}
	if _, err := r.Get(2); err == nil {
		t.Fatalf("expected not found after delete")
	}
}

func TestRelation_UpdateTransform(t *testing.T) {
	r := NewRelation[*e]()
	val := &e{Name: "old"}
	val.SetId(3)
	ins, err := r.Insert(val)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := ins.Confirm(); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	upd, err := r.Update(3, func(orig *e) (*e, error) {
		n := &e{Name: "new"}
		n.SetId(orig.Id())
		return n, nil
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	// ensure new value visible via NewValue
	nv := upd.NewValue()
	if nv.Name != "new" {
		t.Fatalf("expected new name, got %v", nv.Name)
	}
	if err := upd.Confirm(); err != nil {
		t.Fatalf("confirm update: %v", err)
	}
	got, err := r.Get(3)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if got.Name != "new" {
		t.Fatalf("update not applied: %v", got)
	}
}

func TestRelation_Import(t *testing.T) {
	r := NewRelation[*e]()
	items := []*e{}
	for i := 0; i < 5; i++ {
		it := &e{Name: "i"}
		it.SetId(int64(i + 100))
		items = append(items, it)
	}
	if err := r.Import(items); err != nil {
		t.Fatalf("import: %v", err)
	}
	if c := r.Count(); c != 5 {
		t.Fatalf("expected 5 got %d", c)
	}
}

func TestRecord_CommitRollback(t *testing.T) {
	r := NewMutableRecord[*e]()
	// write when uninitialized should succeed
	tst := &e{Name: "x"}
	tst.SetId(999)
	if err := r.Write(tst); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := r.DirtyValue(); err != nil {
		t.Fatalf("dirty value: %v", err)
	}
	if err := r.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if v, err := r.Value(); err != nil || v.Id() != 999 {
		t.Fatalf("value mismatch: %v %v", v, err)
	}
	// delete
	r.Delete()
	if _, err := r.DirtyValue(); err == nil {
		t.Fatalf("expected deleted dirty value error")
	}
}

func TestTransform_GetTransformErrors(t *testing.T) {
	r := NewRelation[*e]()
	val := &e{Name: "x"}
	val.SetId(200)
	ins, err := r.Insert(val)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := ins.Confirm(); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	// transform that changes id
	if _, err := r.Update(200, func(orig *e) (*e, error) {
		n := &e{Name: "z"}
		n.SetId(999)
		return n, nil
	}); err == nil {
		t.Fatalf("expected id changed error")
	}
}
