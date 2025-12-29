package overview

import (
	"strconv"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type topic struct {
	name string
	id   int
}

func (t topic) FilterValue() string {
	return t.name
}

func (t topic) Title() string {
	return t.name
}

func (t topic) Description() string { return strconv.Itoa(t.id) }

type RefreshMsg struct {
	topics []list.Item
}

func RefreshCmd() tea.Cmd {
	return func() tea.Msg {
		return RefreshMsg{
			topics: []list.Item{
				&topic{"topic 1", 1},
				&topic{"topic 2", 2},
				&topic{"topic 3", 3},
				&topic{"topic 4", 4},
				&topic{"topic 5", 5},
			},
		}
	}
}

type OpenChatMsg struct {
	TopicId string
}

func OpenChatCmd(topicId string) tea.Cmd {
	return func() tea.Msg {
		return OpenChatMsg{TopicId: topicId}
	}
}
