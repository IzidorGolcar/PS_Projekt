package login

import (
	"fmt"
	"seminarska/internal/client/components/appbar"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type button int

const (
	login button = iota
	signup
)

type Model struct {
	username textinput.Model
	spinner  spinner.Model
	focused  button
	message  string
	w, h     int
	loading  bool
}

func NewModel() *Model {
	ti := textinput.New()
	ti.Placeholder = "username"
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 30

	sp := spinner.New()
	sp.Spinner = spinner.Points

	return &Model{
		username: ti,
		spinner:  sp,
		focused:  login,
		message:  "Hello!",
		loading:  false,
	}
}

func (l Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, l.spinner.Tick)
}

func (l Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		if l.loading {
			return l, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return l, tea.Quit

		case "tab", "right":
			l.focused = (l.focused + 1) % 2
			return l, nil

		case "left":
			l.focused = (l.focused + 1) % 2
			return l, nil

		case "enter":
			if l.username.Value() == "" {
				l.message = "Username required!"
				return l, nil
			}

			if l.focused == login {
				l.message = fmt.Sprintf("Logging in as %s", l.username.Value())
			} else {
				l.message = fmt.Sprintf("Signing up as %s", l.username.Value())
			}
			l.loading = true
			return l, requestCmd(l.username.Value(), l.focused == signup)
		}

	case tea.WindowSizeMsg:
		l.w = msg.Width
		l.h = msg.Height - appbar.Height

	case spinner.TickMsg:
		l.spinner, cmd = l.spinner.Update(msg)
		return l, cmd

	case ResetMsg:
		l.username.Reset()
		l.message = msg.message
		l.loading = false
		return l, nil
	}

	l.username, cmd = l.username.Update(msg)
	return l, cmd

}

var dialogStyle = lipgloss.NewStyle().
	Align(lipgloss.Center).
	Border(lipgloss.RoundedBorder()).
	Padding(1, 2)

var focusedStyle = lipgloss.NewStyle().
	PaddingLeft(1).
	PaddingRight(1).
	Background(lipgloss.Color("#f7adad")).
	Foreground(lipgloss.Color("#000000")).
	Bold(true)

var unfocusedStyle = lipgloss.NewStyle().
	PaddingLeft(1).
	PaddingRight(1).
	Foreground(lipgloss.Color("#FFFFFF")).
	Background(lipgloss.Color("#909090")).
	Bold(false)

var textStyle = lipgloss.NewStyle().
	PaddingBottom(2)

func (l Model) View() string {
	loginTxt := "Login"
	signupTxt := "Sign Up"

	var loginStyle, signupStyle lipgloss.Style
	if l.focused == login {
		loginStyle = focusedStyle
		signupStyle = unfocusedStyle
	} else {
		loginStyle = unfocusedStyle
		signupStyle = focusedStyle
	}

	var content string
	if l.loading {
		content = l.spinner.View() + " Loading"
	} else {
		content = l.message + "\n\n" +
			textStyle.Render(l.username.View()) + "\n" +
			loginStyle.Render(loginTxt) + "  " + signupStyle.Render(signupTxt)
	}

	return lipgloss.Place(
		l.w,
		l.h,
		lipgloss.Center,
		lipgloss.Center,
		dialogStyle.Render(content),
	)
}
