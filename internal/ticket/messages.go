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

func EditTicket(ticket Ticket, store Store) tea.Cmd {
	return func() tea.Msg {
		model := NewModel(store)
		model.EditTicket(ticket)
		return model
	}
}
