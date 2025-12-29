package input

import (
	"seminarska/internal/client/components/forum/chat/messages"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const Height = 4

type Model struct {
	W     int
	input textinput.Model
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Message ..."
	ti.Focus()
	ti.TextStyle = ti.TextStyle.Background(lipgloss.Color("233"))
	ti.PlaceholderStyle = ti.PlaceholderStyle.Background(lipgloss.Color("233"))
	ti.CharLimit = 50
	ti.Width = 50

	return Model{
		input: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			value := m.input.Value()
			if value != "" {
				m.input.Reset()
				return m, messages.LoadCmd([]messages.Message{{
					Text: value,
					User: "",
				}})
			}
		}
	}

	return m, cmd
}

var style = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder(), true, false, false, false).
	BorderTopForeground(lipgloss.Color("240")).
	Background(lipgloss.Color("233")).
	Padding(1, 2)

func (m Model) View() string {

	content := lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.NewStyle().Background(lipgloss.Color("233")).Width(m.W-9).Render(m.input.View()),
		sendButton(),
	)

	return style.Width(m.W).Render(content)
}

func sendButton() string {
	return lipgloss.
		NewStyle().
		Background(lipgloss.Color("250")).
		Foreground(lipgloss.Color("233")).
		Bold(true).
		Padding(0, 2).
		Render("↪")
}
