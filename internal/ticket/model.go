package ticket

import (
	"fmt"

	"github.com/Kavantix/kantui/internal/messages"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	store Store

	width  int
	height int

	ticket           Ticket
	titleInput       textinput.Model
	descriptionInput textarea.Model
}

func NewModel(store Store) Model {
	titleInput := textinput.New()
	titleInput.Focus()
	titleInput.Width = 1000
	titleInput.Placeholder = "Enter title"

	descriptionInput := textarea.New()
	descriptionInput.Placeholder = "Enter a description"
	descriptionInput.Prompt = ""
	return Model{
		store:            store,
		titleInput:       titleInput,
		descriptionInput: descriptionInput,
	}
}

func (m *Model) EditTicket(ticket Ticket) {
	m.ticket = ticket
	m.titleInput.SetValue(string(ticket.Title))
	m.descriptionInput.SetValue(string(ticket.Description))
}

func (m Model) SetSize(width, height int) tea.Model {
	m.width = width
	m.height = height
	m.titleInput.Width = width
	m.descriptionInput.SetWidth(width)
	m.descriptionInput.SetHeight(height - 2)
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, messages.CloseModal
		case "ctrl+c":
			return m, messages.Quit
		case "enter", "tab":
			if m.titleInput.Focused() {
				m.titleInput.Blur()
				m.descriptionInput.Focus()
				return m, nil
			}
		case "shift+tab":
			if m.descriptionInput.Focused() {
				m.descriptionInput.Blur()
				m.titleInput.Focus()
			}
		case "ctrl+s":
			if m.store == nil {
				return m, func() tea.Msg {
					return messages.CriticalFailureMsg{
						Err: fmt.Errorf("store was not set"),
					}
				}
			}

			var save tea.Cmd
			if m.ticket.ID.IsValid() {
				save = m.store.UpdateTicket(
					m.ticket.ID,
					m.ticket.Status,
					TicketTitle(m.titleInput.Value()),
					TicketDescription(m.descriptionInput.Value()),
				)
			} else {
				save = m.store.New(
					TicketTitle(m.titleInput.Value()),
					TicketDescription(m.descriptionInput.Value()),
				)
			}

			return m, tea.Batch(save, messages.CloseModal)
		}
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.titleInput, cmd = m.titleInput.Update(msg)
	cmds = append(cmds, cmd)
	m.descriptionInput, cmd = m.descriptionInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) OverlayTitle() string {
	if !m.ticket.ID.IsValid() {
		return "New"
	} else {
		return string(m.ticket.Title)
	}
}

var ticketStyle = lipgloss.NewStyle()

func (m Model) View() string {
	if m.titleInput.Focused() {
		m.titleInput.TextStyle = m.descriptionInput.FocusedStyle.Text
		m.titleInput.PromptStyle = m.descriptionInput.FocusedStyle.Text
	} else {
		m.titleInput.TextStyle = m.descriptionInput.BlurredStyle.Text
		m.titleInput.PromptStyle = m.descriptionInput.BlurredStyle.Text
	}
	return ticketStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		m.titleInput.View(),
		"Description",
		m.descriptionInput.View(),
	))

}
