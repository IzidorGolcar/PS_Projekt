package login

import tea "github.com/charmbracelet/bubbletea"

type RequestMsg struct {
	Username string
	NewUser  bool
}

func requestCmd(username string, newUser bool) tea.Cmd {
	return func() tea.Msg {
		return RequestMsg{Username: username, NewUser: newUser}
	}
}

type ResetMsg struct {
	message string
}

func ResetCmd(message string) tea.Cmd {
	return func() tea.Msg {
		return ResetMsg{message: message}
	}
}
