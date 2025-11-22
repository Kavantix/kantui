package ticket

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Kavantix/kantui/internal/database"
	"github.com/Kavantix/kantui/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
)

//go:generate go tool stringerParser -type=Status
type Status uint

type TicketId struct {
	number int64
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
	Load() tea.Msg
	New(title, description string) tea.Cmd
	UpdateStatus(id TicketId, newStatus Status) tea.Cmd
	MoveToNextStatus(id TicketId) tea.Cmd
	MoveToPreviousStatus(id TicketId) tea.Cmd
}

type store struct {
	tickets []Ticket
	queries database.Querier
}

func NewStore(queries database.Querier) Store {
	s := &store{
		queries: queries,
	}
	return s
}

func (s *store) Load() tea.Msg {
	tickets, err := s.queries.GetTickets(context.Background())
	if err != nil {
		return messages.CriticalFailureMsg{
			Err:          err,
			FriendlyText: "Failed to load tickets",
		}
	}
	var status Status
	for _, ticket := range tickets {
		s.tickets = append(s.tickets,
			Ticket{
				ID:          TicketId{ticket.ID},
				Status:      status.Parse(ticket.Status),
				Title:       ticket.Title,
				Description: ticket.Description.String,
			},
		)
	}
	return TicketsUpdatedMsg{Tickets: s.tickets}
}

func (s *store) New(title, description string) tea.Cmd {
	return func() tea.Msg {
		id, err := s.queries.AddTicket(context.Background(), database.AddTicketParams{
			Title: title,
			Description: sql.NullString{
				String: description,
				Valid:  true,
			},
		})
		if err != nil {
			return messages.CriticalFailureMsg{
				Err:          err,
				FriendlyText: "Failed to write new ticket to db",
			}
		}
		s.tickets = append(s.tickets,
			Ticket{
				ID:          TicketId{id},
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
				err := s.queries.UpdateStatus(context.Background(), database.UpdateStatusParams{
					ID:     id.number,
					Status: newStatus.String(),
				})
				if err != nil {
					return messages.CriticalFailureMsg{
						Err:          err,
						FriendlyText: "Failed to update ticket status",
					}
				}
				s.tickets[i].Status = newStatus
			}
		}
		return TicketsUpdatedMsg{s.tickets}
	}

}

func (s *store) MoveToPreviousStatus(id TicketId) tea.Cmd {
	var ticket *Ticket
	for i, t := range s.tickets {
		if t.ID == id {
			ticket = &s.tickets[i]
		}
	}
	if ticket == nil {
		return nil
	}

	var newStatus Status
	switch ticket.Status {
	case Todo:
		return nil
	case InProgress:
		newStatus = Todo
	case Done:
		newStatus = InProgress
	default:
		// assert amount of statusses didnt change
		var _ = [3]any{}[NumberOfStatusses-1]
		panic("unreachable")
	}

	return s.UpdateStatus(id, newStatus)
}

func (s *store) MoveToNextStatus(id TicketId) tea.Cmd {
	var ticket *Ticket
	for i, t := range s.tickets {
		if t.ID == id {
			ticket = &s.tickets[i]
		}
	}
	if ticket == nil {
		return nil
	}

	var newStatus Status
	switch ticket.Status {
	case Todo:
		newStatus = InProgress
	case InProgress:
		newStatus = Done
	case Done:
		return nil
	default:
		// assert amount of statusses didnt change
		var _ = [3]any{}[NumberOfStatusses-1]
		panic("unreachable")
	}

	return s.UpdateStatus(id, newStatus)
}
