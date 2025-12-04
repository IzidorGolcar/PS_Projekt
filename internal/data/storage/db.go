package storage

type Database struct {
	users    *relation[*UserRecord]
	messages *relation[*MessageRecord]
	topics   *relation[*TopicRecord]
}

func NewDatabase() *Database {
	return &Database{
		users:    newRelation[*UserRecord](),
		messages: newRelation[*MessageRecord](),
		topics:   newRelation[*TopicRecord](),
	}
}
