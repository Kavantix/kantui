package app

import (
	"fmt"
	"log/slog"

	"github.com/Kavantix/kantui/internal/column"
	"github.com/Kavantix/kantui/internal/database"
	"github.com/Kavantix/kantui/internal/messages"
	"github.com/Kavantix/kantui/internal/ticket"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	charmansi "github.com/charmbracelet/x/ansi"
	zone "github.com/lrstanley/bubblezone"
)

type Model struct {
	spinner      spinner.Model
	loaded       bool
	windowWidth  int
	windowHeight int
	quitting     bool
	columns      []column.Model
	modals       []ModalModel

	criticalFailure messages.CriticalFailureMsg
}

var _ tea.Model = Model{}

func New() Model {
	m := Model{}
	return m
}

type LoadedMsg struct {
	TicketStore ticket.Store
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		err := database.Migrate()
		if err != nil {
			return messages.CriticalFailureMsg{
				Err:          err,
				FriendlyText: "Failed to migrate database",
			}
		}
		db, err := database.OpenDb()
		if err != nil {
			return messages.CriticalFailureMsg{
				Err:          err,
				FriendlyText: "Failed to open database",
			}
		}
		return LoadedMsg{
			TicketStore: ticket.NewStore(database.New(db)),
		}
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.criticalFailure.Err != nil {
		return m, tea.Quit
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case LoadedMsg:
		m.columns = []column.Model{}
		for _, status := range ticket.Statusses {
			m.columns = append(m.columns, column.New(status, msg.TicketStore))
		}
		m.columns[0].Focus()
		if m.windowWidth > 0 {
			width := m.windowWidth / len(m.columns)
			for _, column := range m.columns {
				column.SetSize(width, m.windowHeight)
			}
		}
		m.loaded = true
		return m, msg.TicketStore.Load
	case messages.CriticalFailureMsg:
		m.criticalFailure = msg
		return m, tea.ExitAltScreen
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		if len(m.columns) > 0 {
			width := msg.Width / len(m.columns)
			for _, column := range m.columns {
				column.SetSize(width, msg.Height)
			}
		}
		return m, nil
	case tea.MouseMsg:
		for i := 0; i < len(m.columns); i++ {
			if zone.Get(fmt.Sprintf("column-%d", i)).InBounds(msg) {
				if msg.Action != tea.MouseActionPress || msg.Button != 1 {
					return m, nil
				}
				for j := 0; j < len(m.columns); j++ {
					m.columns[j].Unfocus()
				}
				m.columns[i].Focus()
				return m, nil
			}
		}
	case ticket.TicketsUpdatedMsg:
		var cmds []tea.Cmd
		for i, _ := range m.columns {
			m.columns[i], cmd = m.columns[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	if len(m.modals) > 0 {
		switch msg.(type) {
		case messages.CloseModalMsg:
			m.modals = m.modals[:len(m.modals)-1]
			return m, nil
		case messages.QuitMsg:
			slog.Info("Quitting")
			m.quitting = true
			return m, tea.Quit
		}
		i := len(m.modals) - 1
		model, cmd := m.modals[i].Update(msg)
		m.modals[i] = model.(ModalModel)
		return m, cmd
	}

	switch msg := msg.(type) {
	case ModalModel:
		m.modals = append(m.modals, msg)
		return m, nil

	case messages.QuitMsg:
		slog.Info("Quitting")
		m.quitting = true
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, messages.Quit
		case "left", "h":
			for i, column := range m.columns {
				if column.Focused() {
					if column.IsCapturingInput() {
						break
					}
					prevIndex := i - 1
					if prevIndex < 0 {
						prevIndex = len(m.columns) - 1
					}
					m.columns[i].Unfocus()
					m.columns[prevIndex].Focus()
					return m, nil
				}
			}
		case "right", "l":
			for i, column := range m.columns {
				if column.Focused() {
					if column.IsCapturingInput() {
						break
					}
					nextIndex := i + 1
					if nextIndex > len(m.columns)-1 {
						nextIndex = 0
					}
					m.columns[i].Unfocus()
					m.columns[nextIndex].Focus()
					return m, nil
				}
			}
		}
	}

	for i, column := range m.columns {
		if column.Focused() {
			m.columns[i], cmd = column.Update(msg)
		}
	}
	return m, cmd

	// slog.Info("Unhandled message: ", slog.String("type", fmt.Sprintf("%T", msg)), slog.Any("msg", msg))
	//
	// return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.criticalFailure.Err != nil {
		style := lipgloss.NewStyle().
			Background(lipgloss.Color("9")).
			Margin(1, 0)

		title := "Failed"
		if m.criticalFailure.FriendlyText != "" {
			title = m.criticalFailure.FriendlyText
		}
		title = style.Render(title)
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			lipgloss.NewStyle().
				Width(m.windowWidth).
				Render(m.criticalFailure.Err.Error()+"\n"),
		)
	}

	if !m.loaded {
		return m.spinner.View()
	}

	columns := []string{}
	for i, column := range m.columns {
		columns = append(columns, zone.Mark(fmt.Sprintf("column-%d", i), column.View()))
	}

	board := lipgloss.JoinHorizontal(
		lipgloss.Center,
		columns...,
	)

	if len(m.modals) > 0 {
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63"))
		verticalMargin := 1
		horizontalMargin := 2

		styleWidth, styleHeight := style.GetFrameSize()
		width, height := m.windowWidth-styleWidth-2*horizontalMargin, m.windowHeight-styleHeight-2*verticalMargin
		modal := m.modals[len(m.modals)-1]
		if modal, ok := modal.(Sizeable[tea.Model]); ok {
			modal.SetSize(width, height)
		}

		overlay := lipgloss.NewStyle().
			Width(width).MaxWidth(width).
			Height(height).MaxHeight(height).
			Render(modal.View())
		overlay = style.Render(overlay)

		return zone.Scan(PlaceOverlay(
			horizontalMargin, verticalMargin,
			overlay, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(charmansi.Strip(board)),
			false,
		))
	} else {
		return zone.Scan(board)
	}
}
