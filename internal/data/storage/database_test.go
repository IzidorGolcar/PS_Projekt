package storage

import (
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"
	"testing"
)

type mockChain struct{}

func (m *mockChain) ForwardRequest(_ *datalink.Record) error {
	return nil
}

func (m *mockChain) SyncedWithTail(_ *datalink.Record) (bool, error) {
	return true, nil
}

func TestDatabase(t *testing.T) {
	chain := &mockChain{}

	t.Run("insert", func(t *testing.T) {
		d := NewDatabase(chain)
		recordA := entities.NewUser("A")
		recordB := entities.NewUser("B")
		err := d.Insert(recordA)
		if err != nil {
			t.Fatal(err)
		}
		err = d.Insert(recordB)
		if err != nil {
			t.Fatal(err)
		}
		if recordA.Id() == recordB.Id() {
			t.Fatal("ids should be different")
		}
	})

	t.Run("duplicate inserts", func(t *testing.T) {
		d := NewDatabase(chain)
		err := d.Insert(entities.NewUser("A"))
		if err != nil {
			t.Fatal(err)
		}
		err = d.Insert(entities.NewUser("B"))
		if err != nil {
			t.Fatal(err)
		}
		err = d.Insert(entities.NewUser("A"))
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("duplicate id", func(t *testing.T) {
		d := NewDatabase(chain)

		recordA := entities.NewUser("A")
		recordB := entities.NewUser("B")

		recordA.SetId(1)
		recordB.SetId(1)

		err := d.Insert(recordA)
		if err != nil {
			t.Fatal(err)
		}

		err = d.Insert(recordB)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get value", func(t *testing.T) {
		d := NewDatabase(chain)
		err := d.Insert(entities.NewUser("A"))
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
