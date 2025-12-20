package entities

type Like struct {
	baseEntity
	UserId    int64 `db:"unique"`
	MessageId int64 `db:"unique"`
}

func NewLike(userId int64, messageId int64) *Like {
	return &Like{UserId: userId, MessageId: messageId}
}
