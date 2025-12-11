package model

//func DatalinkToStorage(dl *datalink.Record) (record storage.Record, err error) {
//	switch p := dl.Payload.(type) {
//	case *datalink.Record_User:
//		record = storage.NewUserRecord(p.User.Name)
//		record.SetId(p.User.Id)
//	case *datalink.Record_Message:
//		record = storage.NewMessageRecord(
//			p.Message.TopicId, p.Message.UserId,
//			p.Message.Text, p.Message.CreatedAt.AsTime(),
//		)
//		record.SetId(p.Message.Id)
//	case *datalink.Record_Like:
//		//request = NewWriteRequest(p.Like)
//	case *datalink.Record_Topic:
//		record = storage.NewTopicRecord(p.Topic.Name)
//		record.SetId(p.Topic.Id)
//	default:
//		err = errors.New("invalid payload")
//	}
//	return
//}
