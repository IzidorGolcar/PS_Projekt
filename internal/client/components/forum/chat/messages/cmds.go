package messages

import tea "github.com/charmbracelet/bubbletea"

type SizeMsg struct {
	H, W int
}

func SizeCmd(h, w int) tea.Cmd {
	return func() tea.Msg {
		return SizeMsg{H: h, W: w}
	}
}

type LoadMsg struct {
	refresh  bool
	Messages []Message
}

func RefreshCmd(messages []Message) tea.Cmd {
	return func() tea.Msg {
		return LoadMsg{Messages: messages, refresh: true}
	}
}

func LoadCmd(messages []Message) tea.Cmd {
	return func() tea.Msg {
		return LoadMsg{Messages: messages}
	}
}
