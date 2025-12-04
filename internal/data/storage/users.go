package storage

type UserRecord struct {
	baseRecord
	name string `storage:"pk"`
}

func NewUserRecord(name string) *UserRecord {
	return &UserRecord{name: name}
}
