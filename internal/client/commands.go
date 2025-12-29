package client

import (
	tea "github.com/charmbracelet/bubbletea"
)

type LoginResultMsg struct {
	success     bool
	explanation string
}

func LoginCommand(username string, newUser bool) tea.Cmd {
	return func() tea.Msg {

		// TODO

		return LoginResultMsg{
			success: true,
		}
	}
}
