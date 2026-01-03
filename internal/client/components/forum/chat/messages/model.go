package messages

import (
	"seminarska/internal/client/components/forum/overview"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Message struct {
	MyMessage bool
	Text      string
	User      string
	Time      time.Time
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
			m.viewport.SetContent(renderMessages(m.messages, m.w))
			m.ready = true
		} else {
			m.viewport.Width = m.w - fx
			m.viewport.Height = m.h - fy
			m.viewport.ScrollDown(m.viewport.TotalLineCount())
		}
	case LoadMsg:
		if equalSlices(msg.Messages, m.messages) || m.topic != msg.Topic {
			return m, LoadRequestCmd(m.topic, 200*time.Millisecond)
		}
		m.messages = msg.Messages
		if m.ready {
			m.viewport.SetContent(renderMessages(m.messages, m.w))
			m.viewport.ScrollDown(m.viewport.TotalLineCount())
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
	background = lipgloss.Color("233")
)

var (
	messageStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Background(background).
			BorderBackground(background).
			Border(lipgloss.RoundedBorder())
	otherMessageStyle = messageStyle.BorderForeground(lipgloss.Color("237"))
	myMessageStyle    = messageStyle.BorderForeground(lipgloss.Color("#f7adad"))
	senderStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
)

func renderMessages(messages []Message, w int) string {
	var msgContents []string
	for _, msg := range messages {
		var style lipgloss.Style
		if msg.MyMessage {
			style = myMessageStyle
		} else {
			style = otherMessageStyle
		}
		content := style.Render(renderContent(msg))
		row := lipgloss.NewStyle().Width(w).Background(background).Render(content)
		msgContents = append(msgContents, row)
	}
	return lipgloss.JoinVertical(lipgloss.Top, msgContents...)
}

func renderContent(message Message) string {
	return message.Text + "\n" + senderStyle.Render(message.User+" • "+message.Time.Local().Format("15:04"))
}

var style = lipgloss.NewStyle().Padding(1, 2)

func (m Model) View() string {
	return style.Height(m.h).Width(m.w).Render(m.viewport.View())
}
