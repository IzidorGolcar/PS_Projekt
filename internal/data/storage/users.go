package storage

type UserRecord struct {
	id   int64
	name string
}

func NewUserRecord(name string) *UserRecord {
	return &UserRecord{name: name}
}

func (r UserRecord) hash() uint64 {
	return stringHash(r.name)
}

func (r UserRecord) Id() RecordId {
	return RecordId(r.id)
}

type Users struct {
	*defaultRelation[UserRecord]
}

func NewUsers() *Users {
	return &Users{newDefaultRelation[UserRecord]()}
}
