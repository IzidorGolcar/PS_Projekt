package messages

import (
	"seminarska/internal/client/components/forum/overview"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Message struct {
	Text string
	User string
	Time time.Time
}

type Model struct {
	w, h     int
	messages []Message
	ready    bool
	viewport viewport.Model
	topic    overview.Topic
}

func NewModel() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SizeMsg:
		fx, fy := style.GetFrameSize()
		m.w, m.h = msg.W, msg.H
		if !m.ready {
			m.viewport = viewport.New(m.w-fx, m.h-fy)
			m.viewport.KeyMap = viewport.KeyMap{}
			m.viewport.SetContent(renderMessages(m.messages))
			m.ready = true
		} else {
			m.viewport.Width = m.w - fx
			m.viewport.Height = m.h - fy
			m.viewport.GotoBottom()
		}
	case LoadMsg:
		if equalSlices(msg.Messages, m.messages) || m.topic != msg.Topic {
			return m, LoadRequestCmd(m.topic, 200*time.Millisecond)
		}
		m.messages = msg.Messages
		if m.ready {
			m.viewport.SetContent(renderMessages(m.messages))
			m.viewport.GotoBottom()
		}
		return m, LoadRequestCmd(msg.Topic, 200*time.Millisecond)
	case overview.SelectTopicMsg:
		m.messages = nil
		m.viewport.SetContent("")
		m.topic = msg.Topic
		return m, LoadRequestCmd(msg.Topic, 0)
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, vpCmd
}

func equalSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

var (
	messageStyle      = lipgloss.NewStyle().Padding(0, 4)
	otherMessageStyle = messageStyle.Border(lipgloss.RoundedBorder()) //.BorderForeground(lipgloss.Color("237")).Background(lipgloss.Color("233")).BorderBackground(lipgloss.Color("233"))
	myMessageStyle    = messageStyle.Border(lipgloss.RoundedBorder()) //.BorderForeground(lipgloss.Color("#f7adad"))
)

func renderMessages(messages []Message) string {
	var msgContents []string
	for _, msg := range messages {
		content := otherMessageStyle.Render(msg.Text)
		msgContents = append(msgContents, content)
	}
	return lipgloss.JoinVertical(lipgloss.Top, msgContents...)
}

var style = lipgloss.NewStyle().Padding(1, 2)

func (m Model) View() string {
	return style.Height(m.h).Width(m.w).Render(m.viewport.View())
}
