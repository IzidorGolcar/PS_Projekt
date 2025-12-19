package storage

import (
	"seminarska/internal/data/storage/entities"
	"seminarska/proto/datalink"
	"seminarska/proto/razpravljalnica"
)

func (d *AppDatabase) GetSnapshot() *datalink.DatabaseSnapshot {
	messages, _ := d.messages.GetAll()
	users, _ := d.users.GetAll()
	topics, _ := d.topics.GetAll()
	likes, _ := d.likes.GetAll()

	snapshot := &datalink.DatabaseSnapshot{
		Users:    make([]*razpravljalnica.User, len(users)),
		Topics:   make([]*razpravljalnica.Topic, len(topics)),
		Messages: make([]*razpravljalnica.Message, len(messages)),
		Likes:    make([]*razpravljalnica.Like, len(likes)),
	}

	for i, message := range messages {
		snapshot.Messages[i] = entities.EntityToDatalink(message).GetMessage()
	}
	for i, like := range likes {
		snapshot.Likes[i] = entities.EntityToDatalink(like).GetLike()
	}
	for i, topic := range topics {
		snapshot.Topics[i] = entities.EntityToDatalink(topic).GetTopic()
	}
	for i, user := range users {
		snapshot.Users[i] = entities.EntityToDatalink(user).GetUser()
	}

	return snapshot

}

func (d *AppDatabase) SetFromSnapshot(snapshot *datalink.DatabaseSnapshot) {
	messages := make([]*entities.Message, len(snapshot.Messages))
	users := make([]*entities.User, len(snapshot.Users))
	topics := make([]*entities.Topic, len(snapshot.Topics))
	likes := make([]*entities.Like, len(snapshot.Likes))

	for i, m := range snapshot.Messages {
		messages[i] = entities.NewMessage(m.UserId, m.TopicId, m.Text, m.CreatedAt.AsTime())
		messages[i].SetId(m.Id)
	}
	for i, u := range snapshot.Users {
		users[i] = entities.NewUser(u.Name)
		users[i].SetId(u.Id)
	}
	for i, t := range snapshot.Topics {
		topics[i] = entities.NewTopic(t.Name)
		topics[i].SetId(t.Id)
	}
	for i, l := range snapshot.Likes {
		likes[i] = entities.NewLike(l.UserId, l.MessageId)
		likes[i].SetId(l.Id)
	}

	err := d.Messages().Import(messages)
	if err != nil {
		panic(err)
	}
	err = d.Users().Import(users)
	if err != nil {
		panic(err)
	}
	err = d.Topics().Import(topics)
	if err != nil {
		panic(err)
	}
	err = d.Likes().Import(likes)
	if err != nil {
		panic(err)
	}
}
