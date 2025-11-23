package ticket

import (
	"fmt"
	"strings"

	"github.com/Kavantix/kantui/internal/confirm"
	"github.com/Kavantix/kantui/internal/messages"
	"github.com/Kavantix/kantui/internal/overlay"
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
	titleInput       *textinput.Model
	descriptionInput textarea.Model
}

// assert
var _ overlay.ModalModel = Model{}
var _ overlay.Sizeable = Model{}

var idStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("69")).
	Bold(true)

func IdStyle() lipgloss.Style {
	return idStyle
}

func NewModel(store Store) Model {
	titleInput := textinput.New()
	titleInput.Focus()
	titleInput.Width = 1000
	titleInput.Placeholder = "New ticket"
	titleInput.Prompt = ""

	descriptionInput := textarea.New()
	descriptionInput.Placeholder = "Enter a description"
	descriptionInput.Prompt = ""
	return Model{
		store:            store,
		titleInput:       &titleInput,
		descriptionInput: descriptionInput,
	}
}

func (m *Model) EditTicket(ticket Ticket) {
	m.ticket = ticket
	m.titleInput.SetValue(string(ticket.Title))
	m.descriptionInput.SetValue(string(ticket.Description))
	m.titleInput.Blur()
	m.descriptionInput.Focus()
}

func (m Model) SetSize(width, height int) overlay.ModalModel {
	const maxWidth = 120
	if width > maxWidth {
		width = maxWidth
	}

	styleWidth, styleHeight := ticketStyle.GetFrameSize()
	m.width = width
	m.height = height
	width -= styleWidth + 2
	height -= styleHeight
	titleWidth := width - 2
	if m.ticket.ID.IsValid() {
		titleWidth = width - 8
	}
	if titleWidth != m.titleInput.Width {
		newTitleInput := textinput.New()
		if m.titleInput.Focused() {
			newTitleInput.Focus()
		}
		newTitleInput.SetValue(m.titleInput.Value())
		newTitleInput.Width = titleWidth
		newTitleInput.Prompt = ""
		newTitleInput.Placeholder = m.titleInput.Placeholder
		m.titleInput = &newTitleInput
	}
	m.descriptionInput.SetWidth(width)
	m.descriptionInput.SetHeight(height - 2)
	return m
}

func (m Model) Size() (width int, height int) {
	return m.width, m.height
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) ticketTitle() TicketTitle {
	value := strings.TrimSpace(m.titleInput.Value())
	return TicketTitle(value)
}

func (m Model) ticketDescription() TicketDescription {
	value := strings.TrimSpace(m.descriptionInput.Value())
	return TicketDescription(value)
}

func (m Model) hasChanged() bool {
	if !m.ticket.ID.IsValid() {
		return m.ticketTitle() != "" || m.ticketDescription() != ""
	}
	return m.ticketTitle() != m.ticket.Title || m.ticketDescription() != m.ticket.Description
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.hasChanged() {
				return m, confirm.Show("Are you sure you want to exit editing?", messages.CloseModal)
			}
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
					m.ticketTitle(),
					m.ticketDescription(),
				)
			} else {
				save = m.store.New(m.ticketTitle(), m.ticketDescription())
			}

			return m, tea.Batch(save, messages.CloseModal)
		}
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	newTitleInput, cmd := m.titleInput.Update(msg)
	m.titleInput = &newTitleInput
	cmds = append(cmds, cmd)
	m.descriptionInput, cmd = m.descriptionInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) modalTitle() string {
	titleBuilder := strings.Builder{}
	if m.ticket.ID.IsValid() {
		titleBuilder.WriteString(idStyle.Render(m.ticket.ID.String()))
		titleBuilder.WriteRune(' ')
	}
	value := m.titleInput.Value()
	maxWidth := lipgloss.Width(value)
	if value == "" {
		maxWidth = lipgloss.Width(m.titleInput.Placeholder)
	}
	if m.titleInput.Focused() {
		maxWidth += 1
	}
	titleBuilder.WriteString(lipgloss.NewStyle().
		MaxWidth(maxWidth).
		Render(m.titleInput.View()))
	return titleBuilder.String()
}

var ticketStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	Padding(1)

func (m Model) View() string {
	if m.titleInput.Focused() {
		m.titleInput.TextStyle = m.descriptionInput.FocusedStyle.Text
		m.titleInput.PromptStyle = m.descriptionInput.FocusedStyle.Text
	} else {
		m.titleInput.TextStyle = m.descriptionInput.BlurredStyle.Text
		m.titleInput.PromptStyle = m.descriptionInput.BlurredStyle.Text
	}
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		"Description",
		m.descriptionInput.View(),
	)
	result := ticketStyle.Render(content)

	return overlay.Place(4, 0, m.modalTitle(), result, false)
}
