package chat

import (
	"math"
	"seminarska/internal/client/components/appbar"
	"seminarska/internal/client/components/forum/chat/input"
	"seminarska/internal/client/components/forum/chat/messages"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	w, h     int
	input    input.Model
	messages messages.Model
}

func NewModel() Model {
	return Model{
		input:    input.NewModel(),
		messages: messages.NewModel(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.messages.Init(), m.input.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	in, inCmd := m.input.Update(msg)
	hist, histCmd := m.messages.Update(msg)
	m.input = in.(input.Model)
	m.messages = hist.(messages.Model)

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.h = msg.Height - appbar.Height
		m.w = msg.Width - int(math.Floor(float64(msg.Width)*0.3))
		m.input.W = m.w
		cmd = messages.SizeCmd(m.h-input.Height, m.w)
	}
	return m, tea.Batch(inCmd, histCmd, cmd)
}

var docStyle = lipgloss.NewStyle().Background(lipgloss.Color("233"))
var contentStyle = lipgloss.NewStyle()

func (m Model) View() string {
	content := contentStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.messages.View(),
			m.input.View(),
		),
	)
	return docStyle.Width(m.w).Height(m.h).Render(content)
}
