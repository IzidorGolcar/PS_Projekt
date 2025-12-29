package messages

import (
	"seminarska/internal/client/components/forum/overview"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type SizeMsg struct {
	H, W int
}

func SizeCmd(h, w int) tea.Cmd {
	return func() tea.Msg {
		return SizeMsg{H: h, W: w}
	}
}

type LoadMsg struct {
	Topic    overview.Topic
	Messages []Message
}

type LoadRequest struct {
	Topic overview.Topic
}

func LoadRequestCmd(topic overview.Topic, delay time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(delay)
		return LoadRequest{Topic: topic}
	}
}
