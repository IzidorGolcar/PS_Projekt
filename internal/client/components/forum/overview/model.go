package overview

import (
	"math"
	"seminarska/internal/client/components/appbar"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Topic struct {
	Name string
	Id   int
}

func (t Topic) FilterValue() string {
	return t.Name
}

func (t Topic) Title() string {
	return t.Name
}

func (t Topic) Description() string { return strconv.Itoa(t.Id) }

type Model struct {
	loaded bool
	list   list.Model
	w, h   int
}

func NewModel() Model {
	var items []list.Item
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.KeyMap = list.KeyMap{
		CursorUp:   key.NewBinding(key.WithKeys("up")),
		CursorDown: key.NewBinding(key.WithKeys("down")),
	}
	l.Title = "Topics"

	return Model{
		loaded: false,
		list:   l,
	}
}

func (m Model) Init() tea.Cmd {
	return LoadRequestCmd(100)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadResponseMsg:
		if len(msg.Topics) == 0 {
			return m, nil
		}
		m.loaded = true
		cmd := m.list.SetItems(msg.listItems())
		return m, tea.Batch(cmd, SelectTopicCmd(msg.Topics[0]))
	case tea.WindowSizeMsg:
		m.w = int(math.Floor(float64(msg.Width) * 0.3))
		m.h = msg.Height - appbar.Height
		x, y := contentStyle.GetFrameSize()
		m.list.SetSize(m.w-x, m.h-y)
	case tea.KeyMsg:
		if msg.String() == "up" || msg.String() == "down" {
			var listCmd tea.Cmd
			m.list, listCmd = m.list.Update(msg)

			if m.list.Items() == nil || len(m.list.Items()) == 0 {
				return m, listCmd
			}
			selectedTopic := m.list.SelectedItem().(Topic)
			return m, tea.Batch(listCmd, SelectTopicCmd(selectedTopic))
		}
	}

	return m, nil
}

var docStyle = lipgloss.NewStyle().Background(lipgloss.Color("234"))
var contentStyle = lipgloss.NewStyle().Margin(1)

func (m Model) View() string {
	content := contentStyle.Render(m.list.View())
	return docStyle.Width(m.w).Height(m.h).Render(content)
}
