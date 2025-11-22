package messages

import (
	tea "github.com/charmbracelet/bubbletea"
)

type CloseModalMsg struct{}

func CloseModal() tea.Msg {
	return CloseModalMsg{}
}

type QuitMsg struct{}

func Quit() tea.Msg {
	return QuitMsg{}
}

type CriticalFailureMsg struct {
	Err          error
	FriendlyText string
}
