package model

import (
	"fmt"
	"log/slog"

	"github.com/Kavantix/kantui/internal/column"
	"github.com/Kavantix/kantui/internal/messages"
	"github.com/Kavantix/kantui/internal/ticket"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	charmansi "github.com/charmbracelet/x/ansi"
	zone "github.com/lrstanley/bubblezone"
)

type Main struct {
	store ticket.Store

	windowWidth  int
	windowHeight int
	quitting     bool
	columns      []column.Model
	modals       []ModalModel
}

var _ tea.Model = Main{}

func New() Main {
	m := Main{
		store: ticket.NewStore(),
	}
	m.columns = []column.Model{}
	for _, status := range ticket.Statusses {
		m.columns = append(m.columns, column.New(status, m.store))
	}
	m.columns[0].Focus()
	return m
}

// Init implements tea.Model.
func (m Main) Init() tea.Cmd {

	return nil
}

// Update implements tea.Model.
func (m Main) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		width := msg.Width / len(m.columns)
		for _, column := range m.columns {
			column.SetSize(width, msg.Height)
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
func (m Main) View() string {
	if m.quitting {
		return "Goodbye!\n"
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
