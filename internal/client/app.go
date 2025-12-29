package client

import (
	"context"
	"fmt"
	"seminarska/internal/client/api"
	"seminarska/internal/client/components"
	"seminarska/internal/client/components/appbar"
	"seminarska/internal/client/components/forum"
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

func (a AppModel) Init() tea.Cmd {
	return components.InitChildren(a.children)
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := components.UpdateChildren(a.children, msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		}

	case login.RequestMsg:
		return a, LoginCommand(msg.Username, msg.NewUser)

	case LoginResultMsg:
		if msg.success {
			a.r = ChatRoute
		} else {
			errMsg := fmt.Sprintf("Login failed: %s", msg.explanation)
			return a, login.ResetCmd(errMsg)
		}
	}
	return a, cmds
}

func (a AppModel) View() string {
	var content string

	switch a.r {
	case LoginRoute:
		content = a.children[1].View()
	case ChatRoute:
		content = a.children[2].View()
	}

	return a.children[0].View() + "\n" + content
}
