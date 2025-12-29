package overview

import (
	"math"
	"seminarska/internal/client/components/appbar"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	focused int
	loaded  bool
	list    list.Model
	w, h    int
}

func NewModel() Model {
	var items []list.Item
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.Title = "Topics"

	return Model{
		loaded: false,
		list:   l,
	}
}

func (m Model) Init() tea.Cmd {
	return RefreshCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case RefreshMsg:
		m.loaded = true
		cmd := m.list.SetItems(msg.topics)
		return m, cmd
	case tea.WindowSizeMsg:
		m.w = int(math.Floor(float64(msg.Width) * 0.3))
		m.h = msg.Height - appbar.Height
		x, y := contentStyle.GetFrameSize()
		m.list.SetSize(m.w-x, m.h-y)
	case tea.KeyMsg:
		if msg.String() == "enter" {
			selected := m.list.SelectedItem().(*topic)
			return m, OpenChatCmd(selected.Title())
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

var docStyle = lipgloss.NewStyle().Background(lipgloss.Color("234"))
var contentStyle = lipgloss.NewStyle().Margin(1)

func (m Model) View() string {
	content := contentStyle.Render(m.list.View())
	return docStyle.Width(m.w).Height(m.h).Render(content)
}
