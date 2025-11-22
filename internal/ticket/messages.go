package ticket

import (
	tea "github.com/charmbracelet/bubbletea"
)

type TicketsUpdatedMsg struct {
	Tickets []Ticket
}

func CreateTicket(store Store) tea.Cmd {
	return func() tea.Msg {
		return NewModel(store)
	}
}
