package ticket

import (
	"fmt"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
)

type Status uint

type TicketId struct {
	number uint32
}

func (i TicketId) IsValid() bool {
	return i.number > 0
}

func (i TicketId) String() string {
	return fmt.Sprintf("TK-%d", i.number)
}

const (
	Todo Status = iota
	InProgress
	Done

	NumberOfStatusses = int(iota)
)

var Statusses = [3]Status{
	Todo, InProgress, Done,
}

func (s Status) ColumnTitle() string {
	switch s {
	case Todo:
		return "TODO"
	case InProgress:
		return "IN PROGRESS"
	case Done:
		return "DONE"
	default:
		// assert amount of statusses didnt change
		var _ = [3]any{}[NumberOfStatusses-1]
		panic("unreachable")
	}
}

type Ticket struct {
	ID          TicketId
	Status      Status
	Title       string
	Description string
}

type Store interface {
	New(title, description string) tea.Cmd
	UpdateStatus(id TicketId, newStatus Status) tea.Cmd
	MoveToNextStatus(id TicketId) tea.Cmd
	MoveToPreviousStatus(id TicketId) tea.Cmd
}

type store struct {
	highestId atomic.Uint32
	tickets   []Ticket
}

func NewStore() Store {
	s := &store{
		highestId: atomic.Uint32{},
	}
	return s
}

func (s *store) New(title, description string) tea.Cmd {
	return func() tea.Msg {
		s.tickets = append(s.tickets,
			Ticket{
				ID:          TicketId{s.highestId.Add(1)},
				Title:       title,
				Description: description,
			},
		)
		return TicketsUpdatedMsg{s.tickets}
	}
}

func (s *store) UpdateStatus(id TicketId, newStatus Status) tea.Cmd {
	return func() tea.Msg {
		for i, t := range s.tickets {
			if t.ID == id {
				s.tickets[i].Status = newStatus
			}
		}
		return TicketsUpdatedMsg{s.tickets}
	}

}

func (s *store) MoveToPreviousStatus(id TicketId) tea.Cmd {
	return func() tea.Msg {
		var ticket *Ticket
		for i, t := range s.tickets {
			if t.ID == id {
				ticket = &s.tickets[i]
			}
		}
		switch ticket.Status {
		case Todo:
			return nil
		case InProgress:
			ticket.Status = Todo
		case Done:
			ticket.Status = InProgress
		default:
			// assert amount of statusses didnt change
			var _ = [3]any{}[NumberOfStatusses-1]
			panic("unreachable")
		}
		return TicketsUpdatedMsg{s.tickets}
	}
}

func (s *store) MoveToNextStatus(id TicketId) tea.Cmd {
	return func() tea.Msg {
		var ticket *Ticket
		for i, t := range s.tickets {
			if t.ID == id {
				ticket = &s.tickets[i]
			}
		}
		switch ticket.Status {
		case Todo:
			ticket.Status = InProgress
		case InProgress:
			ticket.Status = Done
		case Done:
			return nil
		default:
			// assert amount of statusses didnt change
			var _ = [3]any{}[NumberOfStatusses-1]
			panic("unreachable")
		}
		return TicketsUpdatedMsg{s.tickets}
	}
}
