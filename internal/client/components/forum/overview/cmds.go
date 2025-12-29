package overview

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type LoadRequestMsg struct {
	Limit int
}

func LoadRequestCmd(limit int) tea.Cmd {
	return func() tea.Msg {
		return LoadRequestMsg{limit}
	}
}

type LoadResponseMsg struct {
	Success bool
	Err     error
	Topics  []Topic
}

func (m *LoadResponseMsg) listItems() []list.Item {
	items := make([]list.Item, len(m.Topics))
	for i, topic := range m.Topics {
		items[i] = topic
	}
	return items
}

type SelectTopicMsg struct {
	Topic Topic
}

func SelectTopicCmd(topic Topic) tea.Cmd {
	return func() tea.Msg {
		return SelectTopicMsg{topic}
	}
}
