package input

import (
	"seminarska/internal/client/components/forum/overview"

	tea "github.com/charmbracelet/bubbletea"
)

type NewMessageMsg struct {
	Topic overview.Topic
	Text  string
}

func NewMessageCmd(text string, topic overview.Topic) tea.Cmd {
	return func() tea.Msg {
		return NewMessageMsg{
			Text:  text,
			Topic: topic,
		}
	}
}
