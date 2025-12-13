package db

import (
	"seminarska/internal/data/storage/entities"
	"testing"
)

// helper to get the name from an entity via its datalink view
func userName(u *entities.User) string {
	rec := u.ToDatalinkRecord()
	return rec.GetUser().GetName()
}

func TestInsertConfirmAndGetSnapshot(t *testing.T) {
	r := NewRelation[*entities.User]()

	// insert and confirm
	u := entities.NewUser("alice")
	rcpt, err := r.Insert(u)
	if err != nil {
		t.Fatalf("insert error: %v", err)
	}
	if err := rcpt.Confirm(); err != nil {
		t.Fatalf("confirm error: %v", err)
	}

	// fetch snapshot and verify
	snap, err := r.Get(u.Id())
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	val, err := snap.Value()
	if err != nil {
		t.Fatalf("snapshot value error: %v", err)
	}
	if userName(val) != "alice" {
		t.Fatalf("expected name alice, got %q", userName(val))
	}

	// unique constraint on name should reject duplicate
	if _, err := r.Insert(entities.NewUser("alice")); err == nil {
		t.Fatalf("expected unique constraint error on duplicate name, got nil")
	}
}

func TestGetNotFound(t *testing.T) {
	r := NewRelation[*entities.User]()
	if _, err := r.Get(42); err == nil {
		t.Fatalf("expected ErrNotFound, got nil")
	}
}

func TestUpdateConfirm(t *testing.T) {
	r := NewRelation[*entities.User]()

	u := entities.NewUser("alice")
	rcpt, err := r.Insert(u)
	if err != nil {
		t.Fatalf("insert error: %v", err)
	}
	if err := rcpt.Confirm(); err != nil {
		t.Fatalf("confirm error: %v", err)
	}

	upd, err := r.Update(u.Id(), func(e *entities.User) *entities.User {
		// create a new user carrying same id and new name
		nu := entities.NewUser("alice2")
		nu.SetId(e.Id())
		return nu
	})
	if err != nil {
		t.Fatalf("update error: %v", err)
	}

	// Before confirm, Get should still return old value (snapshot isolation)
	snapBefore, err := r.Get(u.Id())
	if err != nil {
		t.Fatalf("get before confirm: %v", err)
	}
	vBefore, err := snapBefore.Value()
	if err != nil {
		t.Fatalf("value before confirm: %v", err)
	}
	if userName(vBefore) != "alice" {
		t.Fatalf("expected alice before confirm, got %q", userName(vBefore))
	}

	if err := upd.Confirm(); err != nil {
		t.Fatalf("confirm update: %v", err)
	}

	// After confirm, new snapshot should show updated value
	snapAfter, err := r.Get(u.Id())
	if err != nil {
		t.Fatalf("get after confirm: %v", err)
	}
	vAfter, err := snapAfter.Value()
	if err != nil {
		t.Fatalf("value after confirm: %v", err)
	}
	if userName(vAfter) != "alice2" {
		t.Fatalf("expected alice2 after confirm, got %q", userName(vAfter))
	}
}

func TestUpdateCancel(t *testing.T) {
	r := NewRelation[*entities.User]()

	u := entities.NewUser("bob")
	rcpt, err := r.Insert(u)
	if err != nil {
		t.Fatalf("insert error: %v", err)
	}
	if err := rcpt.Confirm(); err != nil {
		t.Fatalf("confirm error: %v", err)
	}

	upd, err := r.Update(u.Id(), func(e *entities.User) *entities.User {
		nu := entities.NewUser("bobby")
		nu.SetId(e.Id())
		return nu
	})
	if err != nil {
		t.Fatalf("update error: %v", err)
	}

	// Cancel the update
	upd.Cancel(nil)

	snap, err := r.Get(u.Id())
	if err != nil {
		t.Fatalf("get after cancel: %v", err)
	}
	v, err := snap.Value()
	if err != nil {
		t.Fatalf("value after cancel: %v", err)
	}
	if userName(v) != "bob" {
		t.Fatalf("expected bob after cancel, got %q", userName(v))
	}
}

func TestDeleteCancelRestores(t *testing.T) {
	r := NewRelation[*entities.User]()

	u := entities.NewUser("carol")
	rcpt, err := r.Insert(u)
	if err != nil {
		t.Fatalf("insert error: %v", err)
	}
	if err := rcpt.Confirm(); err != nil {
		t.Fatalf("confirm error: %v", err)
	}

	del, err := r.Delete(u.Id())
	if err != nil {
		t.Fatalf("delete error: %v", err)
	}

	// Cancel the deletion and ensure record remains accessible
	del.Cancel(nil)

	snap, err := r.Get(u.Id())
	if err != nil {
		t.Fatalf("get after delete cancel: %v", err)
	}
	v, err := snap.Value()
	if err != nil {
		t.Fatalf("value after delete cancel: %v", err)
	}
	if userName(v) != "carol" {
		t.Fatalf("expected carol after delete cancel, got %q", userName(v))
	}
}

func TestUniqueConstraintOnInsert(t *testing.T) {
	r := NewRelation[*entities.User]()

	// Insert and confirm first user
	u1 := entities.NewUser("dave")
	rcpt1, err := r.Insert(u1)
	if err != nil {
		t.Fatalf("insert error: %v", err)
	}
	if err := rcpt1.Confirm(); err != nil {
		t.Fatalf("confirm error: %v", err)
	}

	// Attempt to insert duplicate by unique name
	if _, err := r.Insert(entities.NewUser("dave")); err == nil {
		t.Fatalf("expected unique constraint error for duplicate name, got nil")
	}
}

func TestDeleteConfirm(t *testing.T) {
	r := NewRelation[*entities.User]()

	// Insert and confirm a user
	u := entities.NewUser("erin")
	rcpt, err := r.Insert(u)
	if err != nil {
		t.Fatalf("insert error: %v", err)
	}
	if err := rcpt.Confirm(); err != nil {
		t.Fatalf("confirm error: %v", err)
	}

	// Delete and confirm
	del, err := r.Delete(u.Id())
	if err != nil {
		t.Fatalf("delete error: %v", err)
	}
	if err := del.Confirm(); err != nil {
		t.Fatalf("confirm delete: %v", err)
	}

	// After confirm, the record should not be found
	if _, err := r.Get(u.Id()); err == nil {
		t.Fatalf("expected ErrNotFound after delete confirm, got nil")
	}

	// Unique index should be freed so we can insert the same name again
	if _, err := r.Insert(entities.NewUser("erin")); err != nil {
		t.Fatalf("reinsert after delete should succeed, got error: %v", err)
	}
}
