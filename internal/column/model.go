package column

import (
	"github.com/Kavantix/kantui/internal/confirm"
	"github.com/Kavantix/kantui/internal/ticket"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	store ticket.Store

	delegate *list.DefaultDelegate

	status  ticket.Status
	focused bool
	list    *list.Model
}

type item struct {
	ticket ticket.Ticket
}

var idStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("63")).
	Bold(true)

func (i item) Title() string {
	return string(i.ticket.Title) + " " + idStyle.Render(i.ticket.ID.String())
}
func (i item) Description() string { return string(i.ticket.Description) }
func (i item) FilterValue() string { return string(i.ticket.Title) }

var defaultStyles = list.NewDefaultItemStyles()

func New(status ticket.Status, store ticket.Store) Model {
	delegate := list.NewDefaultDelegate()
	listModel := list.New(
		[]list.Item{},
		&delegate, 0, 0,
	)
	listModel.SetShowHelp(false)
	listModel.Title = status.ColumnTitle()
	switch status {
	case ticket.InProgress:
		listModel.Styles.Title = listModel.Styles.Title.Background(lipgloss.Color("4"))
	case ticket.Done:
		listModel.Styles.Title = listModel.Styles.Title.Background(lipgloss.Color("2"))
	}
	m := Model{
		delegate: &delegate,
		store:    store,
		status:   status,
		list:     &listModel,
	}
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Focus() {
	m.focused = true
}

func (m *Model) Unfocus() {
	m.focused = false
}

func (m Model) Focused() bool {
	return m.focused
}

func (m Model) IsCapturingInput() bool {
	return m.list.SettingFilter()
}

var (
	style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder())
)

func (m Model) SetSize(width, height int) {
	styleX, styleY := style.GetFrameSize()
	m.list.SetSize(width-styleX, height-styleY)
}

func (m Model) setTickets(tickets []ticket.Ticket) tea.Cmd {
	var items []list.Item
	for _, ticket := range tickets {
		if ticket.Status == m.status {
			items = append(items, item{ticket: ticket})
		}
	}
	return m.list.SetItems(items)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ticket.TicketsUpdatedMsg:
		return m, m.setTickets(msg.Tickets)
	case tea.KeyMsg:
		if m.IsCapturingInput() {
			break
		}
		switch msg.String() {
		case "esc":
			return m, nil
		case "n":
			return m, ticket.CreateTicket(m.store)
		case "d":
			item, ok := m.list.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			return m, confirm.Show("Are you sure you want to delete?", m.store.DeleteTicket(item.ticket.ID))
		case "e":
			item, ok := m.list.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			return m, ticket.EditTicket(item.ticket, m.store)
		case "b":
			item, ok := m.list.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			return m, m.store.MoveToPreviousStatus(item.ticket.ID)
		case " ":
			item, ok := m.list.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			return m, m.store.MoveToNextStatus(item.ticket.ID)
		}
	}
	newListModel, cmd := m.list.Update(msg)
	m.list = &newListModel
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	borderColor := style.GetBorderTopForeground()
	if m.focused {
		borderColor = lipgloss.Color("63")
		m.delegate.Styles.SelectedTitle = defaultStyles.SelectedTitle
		m.delegate.Styles.SelectedDesc = defaultStyles.SelectedDesc
	} else {
		m.delegate.Styles.SelectedTitle = defaultStyles.NormalTitle
		m.delegate.Styles.SelectedDesc = defaultStyles.NormalDesc
	}
	return style.
		Width(m.list.Width()).
		Height(m.list.Height()).
		BorderForeground(borderColor).
		Render(m.list.View())
}
