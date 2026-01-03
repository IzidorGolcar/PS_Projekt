package client

import (
	"log"
	"seminarska/internal/client/components/forum/chat/messages"
	"seminarska/internal/client/components/forum/overview"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"google.golang.org/grpc/status"
)

type LoginResultMsg struct {
	success     bool
	explanation string
}

func (m AppModel) LoginCommand(username string, newUser bool) tea.Cmd {
	return func() tea.Msg {
		if newUser {
			err := m.client.SignUp(username)
			if err == nil {
				return LoginResultMsg{
					success: true,
				}
			}
			log.Println("failed to sign up:", err)
			s := status.Convert(err)
			return LoginResultMsg{
				success:     false,
				explanation: s.Message(),
			}
		}

		err := m.client.Login(username)
		if err == nil {
			return LoginResultMsg{
				success: true,
			}
		}
		log.Println("failed to login:", err)
		s := status.Convert(err)
		return LoginResultMsg{
			success:     false,
			explanation: s.Message(),
		}
	}
}

func (m AppModel) LoadResponseCmd() tea.Cmd {
	return func() tea.Msg {
		topics, err := m.client.ListTopics()
		if err != nil {
			log.Println("failed to load topics:", err)
			return overview.LoadResponseMsg{
				Success: false,
				Err:     err,
			}
		}

		items := make([]overview.Topic, len(topics))
		for i, topic := range topics {
			items[i] = overview.Topic{Name: topic.GetName(), Id: int(topic.GetId())}
		}
		return overview.LoadResponseMsg{
			Success: true,
			Topics:  items,
		}
	}
}

func (m AppModel) LoadMsgCmd(topic overview.Topic) tea.Cmd {
	return func() tea.Msg {
		res, err := m.client.GetMessages(topic.Id)
		if err != nil {
			log.Println("failed to subscribe:", err)
			return nil
		}

		items := make([]messages.Message, len(res))
		for i, msg := range res {
			username, err := m.client.GetUsername(int(msg.GetUserId()))
			if err != nil {
				log.Println("failed to get username:", err)
			}

			items[i] = messages.Message{
				MyMessage: msg.GetUserId() == int64(m.client.UserId()),
				Text:      msg.GetText(),
				User:      username,
				Time:      msg.GetCreatedAt().AsTime(),
			}
		}

		sort.Slice(items, func(i, j int) bool {
			return items[i].Time.Before(items[j].Time)
		})

		return messages.LoadMsg{Messages: items, Topic: topic}
	}
}

func (m AppModel) SendMessageCmd(topic overview.Topic, text string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.PostMessage(topic.Id, text)
		if err != nil {
			return nil
		}
		return m.LoadMsgCmd(topic)()
	}
}
