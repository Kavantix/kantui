package column

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Kavantix/kantui/internal/confirm"
	"github.com/Kavantix/kantui/internal/ticket"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

type Model struct {
	store ticket.Store

	delegate *listDelegate

	status  ticket.Status
	focused bool
	list    *list.Model

	lastClick *struct {
		ticketId ticket.TicketId
		at       time.Time
	}
}

type item struct {
	ticket ticket.Ticket
}

func (i item) Title() string {
	return string(i.ticket.Title) + " " + i.ticket.ID.String()
}
func (i item) Description() string { return string(i.ticket.Description) }
func (i item) FilterValue() string { return string(i.ticket.Title) + " " + i.ticket.ID.String() }

var defaultStyles = list.NewDefaultItemStyles()

type listDelegate struct {
	list.DefaultDelegate
	width int
}

func (d listDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	buffer := strings.Builder{}
	d.DefaultDelegate.Render(&buffer, m, index, listItem)
	id := listItem.(item).ticket.ID.String()
	content := buffer.String()
	content = strings.Replace(content, id, ticket.IdStyle().Render(id), 1)
	fmt.Fprint(w, zone.Mark(id, lipgloss.NewStyle().Width(d.width).Render(content)))
}

func New(status ticket.Status, store ticket.Store) Model {
	delegate := listDelegate{list.NewDefaultDelegate(), 0}
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
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("239"))
)

func (m Model) SetSize(width, height int) {
	styleX, styleY := style.GetFrameSize()
	m.list.SetSize(width-styleX, height-styleY)
	m.delegate.width = width - styleX
}

func (m Model) setTickets(tickets []ticket.Ticket) tea.Cmd {
	var selectedTicketId ticket.TicketId
	visibleItems := m.list.VisibleItems()
	selectedIndex := m.list.Index()
	if len(visibleItems) > 0 {
		if selectedIndex >= 0 && selectedIndex < len(visibleItems) {
			ticket := visibleItems[selectedIndex].(item).ticket
			selectedTicketId = ticket.ID
		}
	}

	var items []list.Item
	var newSelectedIndex = selectedIndex
	for _, ticket := range tickets {
		if ticket.Status == m.status {
			if ticket.ID == selectedTicketId {
				newSelectedIndex = len(items)
			}
			items = append(items, item{ticket: ticket})
		}
	}
	cmd := m.list.SetItems(items)
	if newSelectedIndex != selectedIndex {
		m.list.Select(newSelectedIndex)
	}
	return cmd
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ticket.TicketsUpdatedMsg:
		return m, m.setTickets(msg.Tickets)
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			newListModel := *m.list
			var cmd tea.Cmd
			switch msg.Button {
			case tea.MouseButtonWheelDown:
				newListModel, cmd = m.list.Update(tea.KeyMsg{Type: tea.KeyDown})
			case tea.MouseButtonWheelUp:
				newListModel, cmd = m.list.Update(tea.KeyMsg{Type: tea.KeyUp})
			case tea.MouseButtonLeft:
				visibleItems := newListModel.VisibleItems()
				for i, listItem := range visibleItems {
					item := listItem.(item)
					if zone.Get(item.ticket.ID.String()).InBounds(msg) {
						newListModel.Select(i)
						if m.lastClick != nil &&
							m.lastClick.ticketId == item.ticket.ID &&
							time.Since(m.lastClick.at) < 500*time.Millisecond {
							m.lastClick = nil
							return m, ticket.EditTicket(item.ticket, m.store)
						} else {
							m.lastClick = &struct {
								ticketId ticket.TicketId
								at       time.Time
							}{
								item.ticket.ID, time.Now(),
							}
						}
						break
					}
				}
			}
			m.list = &newListModel
			return m, cmd
		}
	case tea.KeyMsg:
		if m.IsCapturingInput() {
			break
		}
		switch msg.String() {
		case "esc":
			if m.list.IsFiltered() {
				m.list.ResetFilter()
			}
			return m, nil
		case "c":
			return m, ticket.CreateTicket(m.store)
		case "d":
			item, ok := m.list.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			msgBuilder := strings.Builder{}
			msgBuilder.WriteString("Are you sure you want to delete ")
			msgBuilder.WriteString(ticket.IdStyle().Render(item.ticket.ID.String()))
			msgBuilder.WriteRune('?')
			return m, confirm.Show(msgBuilder.String(), m.store.DeleteTicket(item.ticket.ID))
		case "e", " ":
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
		case "n":
			item, ok := m.list.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			return m, m.store.MoveToNextStatus(item.ticket.ID)
		case "J", "shift+down":
			visibleItems := m.list.VisibleItems()
			index := m.list.Index()
			newIndex := index + 1
			return m.rankDown(index, newIndex, visibleItems)
		case "B":
			visibleItems := m.list.VisibleItems()
			index := m.list.Index()
			newIndex := len(visibleItems) - 1
			return m.rankDown(index, newIndex, visibleItems)
		case "K", "shift+up":
			visibleItems := m.list.VisibleItems()
			index := m.list.Index()
			newIndex := index - 1
			return m.rankUp(index, newIndex, visibleItems)
		case "T":
			visibleItems := m.list.VisibleItems()
			index := m.list.Index()
			newIndex := 0
			return m.rankUp(index, newIndex, visibleItems)
		}
	}
	newListModel, cmd := m.list.Update(msg)
	m.list = &newListModel
	return m, cmd
}

func (m Model) rankUp(index int, newIndex int, visibleItems []list.Item) (Model, tea.Cmd) {
	if index == newIndex || newIndex < 0 || index > len(visibleItems)-1 {
		return m, nil
	}
	ticket := visibleItems[index].(item).ticket
	previousTicket := visibleItems[newIndex].(item).ticket
	return m, m.store.RankTicketBeforeTicket(ticket.ID, previousTicket.ID)
}

func (m Model) rankDown(index int, newIndex int, visibleItems []list.Item) (Model, tea.Cmd) {
	if index == newIndex || index < 0 || newIndex > len(visibleItems)-1 {
		return m, nil
	}
	ticket := visibleItems[index].(item).ticket
	nextTicket := visibleItems[newIndex].(item).ticket
	return m, m.store.RankTicketAfterTicket(ticket.ID, nextTicket.ID)
}

// View implements tea.Model.
func (m Model) View() string {
	borderColor := style.GetBorderTopForeground()
	if m.focused {
		borderColor = lipgloss.Color("13")
		m.delegate.Styles.SelectedTitle = defaultStyles.SelectedTitle
		m.delegate.Styles.SelectedDesc = defaultStyles.SelectedDesc
	} else {
		m.delegate.Styles.SelectedTitle = defaultStyles.SelectedTitle.
			BorderForeground(lipgloss.Color("33")).
			Foreground(lipgloss.Color("248"))

		m.delegate.Styles.SelectedDesc = defaultStyles.SelectedDesc.
			BorderForeground(lipgloss.Color("68")).
			Foreground(defaultStyles.NormalDesc.GetForeground())
		m.delegate.Styles.NormalTitle = defaultStyles.NormalTitle.Foreground(lipgloss.Color("248"))
	}
	return style.
		Width(m.list.Width()).
		Height(m.list.Height()).
		BorderForeground(borderColor).
		Render(m.list.View())
}
