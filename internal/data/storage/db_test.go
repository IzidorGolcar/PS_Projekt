package storage

import (
	"testing"
)

func TestDatabase(t *testing.T) {
	t.Run("duplicate inserts", func(t *testing.T) {
		d := NewDatabase()
		rec1, err := d.Insert(NewUserRecord("A"))
		if err != nil {
			t.Fatal(err)
		}
		rec2, err := d.Insert(NewUserRecord("B"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = d.Insert(NewUserRecord("A"))
		if err == nil {
			t.Fatal("expected error")
		}
		err = rec1.Confirm(0)
		if err != nil {
			t.Fatal(err)
		}
		err = rec2.Confirm(0)
		if err == nil {
			t.Fatal("expected error: writing record with duplicate id")
		}

		if len(d.users.pending) != 0 {
			t.FailNow()
		}
		if len(d.users.confirmed) != 1 {
			t.FailNow()
		}
	})

	t.Run("get value", func(t *testing.T) {
		d := NewDatabase()
		rec, err := d.Insert(NewUserRecord("A"))
		if err != nil {
			t.Fatal(err)
		}
		err = rec.Confirm(0)
		if err != nil {
			t.Fatal(err)
		}
		record, err := d.GetUser(0)
		if err != nil {
			t.Fatal(err)
		}
		if record.name != "A" {
			t.FailNow()
		}
	})
}
