package overview

import (
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type LoadRequestMsg struct {
	Limit int
}

func LoadRequestCmd(limit int, delay time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(delay)
		return LoadRequestMsg{limit}
	}
}

type LoadResponseMsg struct {
	Success bool
	Err     error
	Topics  []Topic
}

func (m *LoadResponseMsg) listItems() []list.Item {
	slices.SortFunc(m.Topics, func(a, b Topic) int {
		return a.Id - b.Id
	})
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
