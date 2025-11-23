package ticket

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

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
	if i.number <= 0 {
		return "INVALID"
	}
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

type TicketTitle string
type TicketDescription string

type Ticket struct {
	ID          TicketId
	rank        int64
	Status      Status
	Title       TicketTitle
	Description TicketDescription
}

type Store interface {
	Load() tea.Msg
	New(title TicketTitle, description TicketDescription) tea.Cmd
	UpdateTicket(id TicketId, newTitle TicketTitle, newDescription TicketDescription) tea.Cmd
	UpdateStatus(id TicketId, newStatus Status) tea.Cmd
	RankTicketAfterTicket(id, afterId TicketId) tea.Cmd
	RankTicketBeforeTicket(id, beforeId TicketId) tea.Cmd
	MoveToNextStatus(id TicketId) tea.Cmd
	MoveToPreviousStatus(id TicketId) tea.Cmd
	DeleteTicket(id TicketId) tea.Cmd
}

type store struct {
	tickets []Ticket
	db      database.Connection
}

func NewStore(db database.Connection) Store {
	s := &store{
		db: db,
	}
	return s
}

func (s *store) Load() tea.Msg {
	tickets, err := s.db.GetTickets(context.Background())
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
				rank:        ticket.Rank,
				Title:       TicketTitle(ticket.Title),
				Description: TicketDescription(ticket.Description.String),
			},
		)
	}
	return TicketsUpdatedMsg{Tickets: s.tickets}
}

func (s *store) New(title TicketTitle, description TicketDescription) tea.Cmd {
	return func() tea.Msg {
		row, err := s.db.AddTicket(context.Background(), database.AddTicketParams{
			Title: string(title),
			Description: sql.NullString{
				String: string(description),
				Valid:  description != "",
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
				ID:          TicketId{row.ID},
				rank:        row.Rank,
				Title:       title,
				Description: description,
			},
		)
		return TicketsUpdatedMsg{s.tickets}
	}
}

func (s *store) UpdateTicket(id TicketId, newTitle TicketTitle, newDescription TicketDescription) tea.Cmd {
	return func() tea.Msg {
		for i, t := range s.tickets {
			if t.ID == id {
				err := s.db.UpdateTicketContent(context.Background(), database.UpdateTicketContentParams{
					ID:    id.number,
					Title: string(newTitle),
					Description: sql.NullString{
						String: string(newDescription),
						Valid:  newDescription != "",
					},
				})
				if err != nil {
					return messages.CriticalFailureMsg{
						Err:          err,
						FriendlyText: "Failed to update ticket",
					}
				}
				s.tickets[i].Title = newTitle
				s.tickets[i].Description = newDescription
			}
		}
		return TicketsUpdatedMsg{s.tickets}
	}

}

func (s *store) UpdateStatus(id TicketId, newStatus Status) tea.Cmd {
	return func() tea.Msg {
		for i, t := range s.tickets {
			if t.ID == id {
				err := s.db.UpdateStatus(context.Background(), database.UpdateStatusParams{
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

func (s *store) RankTicketAfterTicket(id, afterId TicketId) tea.Cmd {
	return func() tea.Msg {
		index := s.indexOfTicket(id)
		afterIndex := s.indexOfTicket(afterId)
		if afterIndex < 0 || index < 0 {
			return nil
		}
		if index >= afterIndex {
			return messages.CriticalFailureMsg{
				Err: fmt.Errorf("Ranking after should only be used when ticket is not already after other ticket"),
			}
		}
		return s.rankTicket(index, afterIndex+1)
	}
}

func (s *store) RankTicketBeforeTicket(id, beforeId TicketId) tea.Cmd {
	return func() tea.Msg {
		index := s.indexOfTicket(id)
		beforeIndex := s.indexOfTicket(beforeId)
		if beforeIndex < 0 || index < 0 {
			return nil
		}
		if index <= beforeIndex {
			return messages.CriticalFailureMsg{
				Err: fmt.Errorf("Ranking before should only be used when ticket is not already before other ticket"),
			}
		}
		return s.rankTicket(index, beforeIndex)
	}
}

func (s *store) indexOfTicket(id TicketId) int {
	return slices.IndexFunc(s.tickets, func(ticket Ticket) bool { return ticket.ID == id })
}

func (s *store) rankTicket(currentIndex, newIndex int) tea.Msg {
	if currentIndex == newIndex {
		return nil
	}

	newRank, err := s.computeNewRank(currentIndex, newIndex)
	if err != nil {
		return messages.CriticalFailureMsg{
			Err:          err,
			FriendlyText: "Failed to update rank",
		}
	}

	ticket := s.tickets[currentIndex]

	// Update database
	err = s.db.UpdateRank(context.Background(), database.UpdateRankParams{
		ID:   ticket.ID.number,
		Rank: newRank,
	})
	if err != nil {
		return messages.CriticalFailureMsg{
			Err:          err,
			FriendlyText: "Failed to update rank",
		}
	}

	// Update loaded tickets
	ticket.rank = newRank
	if newIndex > currentIndex {
		s.tickets = slices.Insert(s.tickets, newIndex, ticket)
		s.tickets = slices.Delete(s.tickets, currentIndex, currentIndex+1)
	} else {
		s.tickets = slices.Delete(s.tickets, currentIndex, currentIndex+1)
		s.tickets = slices.Insert(s.tickets, newIndex, ticket)
	}

	return TicketsUpdatedMsg{s.tickets}
}

func (s *store) computeNewRank(currentIndex, newIndex int) (int64, error) {
	if currentIndex == newIndex {
		return 0, errors.New("current and new index are the same")
	}
	if currentIndex < 0 || currentIndex >= len(s.tickets) || newIndex < 0 || newIndex > len(s.tickets) {
		return 0, fmt.Errorf("invalid indices: currentIndex=%d, newIndex=%d", currentIndex, newIndex)
	}

	// Determine the ticket to compare to
	var comparisonTicket Ticket
	if newIndex > currentIndex {
		comparisonTicket = s.tickets[newIndex-1]
	} else {
		comparisonTicket = s.tickets[newIndex]
	}

	// Compute new rank
	var newRank int64
	if newIndex == 0 {
		newRank = comparisonTicket.rank - 1_000_000
	} else if newIndex >= len(s.tickets) {
		newRank = comparisonTicket.rank + 1_000_000
	} else {
		var gap int64
		if newIndex > currentIndex {
			gap = s.tickets[newIndex].rank - comparisonTicket.rank
			newRank = comparisonTicket.rank + gap/2
		} else {
			gap = comparisonTicket.rank - s.tickets[newIndex-1].rank
			newRank = comparisonTicket.rank - gap/2
		}

		if gap <= 1 {
			return 0, fmt.Errorf("ran out of room between tickets")
		}
	}

	return newRank, nil
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

func (s *store) DeleteTicket(id TicketId) tea.Cmd {
	return func() tea.Msg {
		err := s.db.DeleteTicket(context.Background(), id.number)
		if err != nil {
			return messages.CriticalFailureMsg{
				Err:          err,
				FriendlyText: "Failed to delete ticket",
			}
		}
		s.tickets = slices.DeleteFunc(s.tickets, func(ticket Ticket) bool {
			return ticket.ID == id
		})
		return TicketsUpdatedMsg{s.tickets}
	}
}
