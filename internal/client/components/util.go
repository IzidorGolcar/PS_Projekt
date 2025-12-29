package components

import tea "github.com/charmbracelet/bubbletea"

func InitChildren(children []tea.Model) tea.Cmd {
	cmds := make([]tea.Cmd, len(children))
	for i, c := range children {
		cmds[i] = c.Init()
	}
	return tea.Batch(cmds...)
}

func UpdateChildren(children []tea.Model, msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(children))
	for i, c := range children {
		children[i], cmds[i] = c.Update(msg)
	}
	return tea.Batch(cmds...)
}
