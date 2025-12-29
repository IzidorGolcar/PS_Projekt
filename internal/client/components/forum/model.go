package forum

import (
	"seminarska/internal/client/components/forum/chat"
	"seminarska/internal/client/components/forum/overview"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	openedTopic int
	overview    overview.Model
	messages    chat.Model
	w, h        int
}

func NewModel() *Model {
	return &Model{
		overview: overview.NewModel(),
		messages: chat.NewModel(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.overview.Init(), m.messages.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ovw, ovwCmd := m.overview.Update(msg)
	mssg, mssgCmd := m.messages.Update(msg)
	m.overview = ovw.(overview.Model)
	m.messages = mssg.(chat.Model)
	return m, tea.Batch(ovwCmd, mssgCmd)
}

func (m Model) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.overview.View(),
		m.messages.View(),
	)
}
