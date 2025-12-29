package appbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#000000")).
	Background(lipgloss.Color("#f7adad")).
	PaddingTop(1).
	PaddingBottom(1).
	PaddingLeft(4).
	PaddingRight(4).
	Align(lipgloss.Center)

type Model struct {
	w     int
	title string
}

const Height = 3

func NewModel(title string) *Model {
	return &Model{title: title}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w = msg.Width
	}
	return m, nil
}

func (m Model) View() string {
	return style.Width(m.w).Render(m.title)
}
