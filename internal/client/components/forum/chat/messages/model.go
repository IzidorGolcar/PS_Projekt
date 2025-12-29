package messages

import (
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
}

func NewModel() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return LoadCmd([]Message{
		{
			Text: "Hello picka!",
			User: "kurac",
			Time: time.Now(),
		},
	})
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
		if msg.refresh {
			m.messages = msg.Messages
		} else {
			m.messages = append(m.messages, msg.Messages...)
		}
		if m.ready {
			m.viewport.SetContent(renderMessages(m.messages))
			m.viewport.GotoBottom()
		}
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, vpCmd
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
