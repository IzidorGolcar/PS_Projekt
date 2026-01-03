package client

import (
	"context"
	"fmt"
	"seminarska/internal/client/api"
	"seminarska/internal/client/components"
	"seminarska/internal/client/components/appbar"
	"seminarska/internal/client/components/forum"
	"seminarska/internal/client/components/forum/chat/input"
	"seminarska/internal/client/components/forum/chat/messages"
	"seminarska/internal/client/components/forum/overview"
	"seminarska/internal/client/components/login"

	tea "github.com/charmbracelet/bubbletea"
)

//go:generate stringer -type=route

type route int

const (
	LoginRoute route = iota
	ChatRoute
)

type AppModel struct {
	client   *api.Client
	children []tea.Model
	r        route
}

func NewAppModel(controlAddr string) AppModel {
	return AppModel{
		children: []tea.Model{
			appbar.NewModel("Razpravljalnica"),
			login.NewModel(),
			forum.NewModel(),
		},
		client: api.NewClient(context.Background(), controlAddr),
		r:      LoginRoute,
	}
}

func (m AppModel) Init() tea.Cmd {
	return components.InitChildren(m.children)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		var cmd tea.Cmd
		switch m.r {
		case LoginRoute:
			m.children[1], cmd = m.children[1].Update(msg)
		case ChatRoute:
			m.children[2], cmd = m.children[2].Update(msg)
		}
		cmds = append(cmds, cmd)

	case tea.WindowSizeMsg:
		cmds = append(cmds, components.UpdateChildren(m.children, msg))

	default:
		cmds = append(cmds, components.UpdateChildren(m.children, msg))
	}

	switch msg := msg.(type) {
	case login.RequestMsg:
		return m, tea.Batch(tea.Batch(cmds...), m.LoginCommand(msg.Username, msg.NewUser))

	case LoginResultMsg:
		if msg.success {
			m.r = ChatRoute
		} else {
			errMsg := fmt.Sprintf("Login failed: %s", msg.explanation)
			return m, tea.Batch(tea.Batch(cmds...), login.ResetCmd(errMsg))
		}

	case overview.LoadRequestMsg:
		return m, tea.Batch(tea.Batch(cmds...), m.LoadResponseCmd())

	case messages.LoadRequest:
		return m, tea.Batch(tea.Batch(cmds...), m.LoadMsgCmd(msg.Topic))

	case input.NewMessageMsg:
		return m, tea.Batch(tea.Batch(cmds...), m.SendMessageCmd(msg.Topic, msg.Text))
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	var content string

	switch m.r {
	case LoginRoute:
		content = m.children[1].View()
	case ChatRoute:
		content = m.children[2].View()
	}

	return m.children[0].View() + "\n" + content
}
