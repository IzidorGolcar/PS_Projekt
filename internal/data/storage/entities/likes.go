package entities

type Like struct {
	baseEntity
	userId    int64 `db:"unique"`
	messageId int64 `db:"unique"`
}

func NewLike(userId int64, messageId int64) *Like {
	return &Like{userId: userId, messageId: messageId}
}

func (l *Like) UserId() int64 {
	return l.userId
}

func (l *Like) MessageId() int64 {
	return l.messageId
}
