package storage

import (
	"errors"
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

		err = rec1.Confirm()
		if err != nil {
			t.Fatal(err)
		}
		err = rec2.Confirm()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("receipts", func(t *testing.T) {
		d := NewDatabase()
		rec1, err := d.Insert(NewUserRecord("A"))
		if err != nil {
			t.Fatal(err)
		}
		rec2, err := d.Insert(NewUserRecord("B"))
		if err != nil {
			t.Fatal(err)
		}
		rec3, err := d.Insert(NewUserRecord("C"))
		if err != nil {
			t.Fatal(err)
		}

		err = rec1.Confirm()
		if err != nil {
			t.Fatal(err)
		}
		rec2.Cancel(errors.New("canceled"))
		err = rec3.Confirm()
		if err != nil {
			t.Fatal(err)
		}

		if len(d.users.confirmed) != 2 {
			t.Fatal("expected 2 confirmed records")
		}
		if len(d.users.pending) != 0 {
			t.Fatal("expected 0 pending records")
		}

	})

	t.Run("get value", func(t *testing.T) {
		d := NewDatabase()
		rec, err := d.Insert(NewUserRecord("A"))
		if err != nil {
			t.Fatal(err)
		}
		err = rec.Confirm()
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
