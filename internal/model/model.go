package model

import (
	"log/slog"

	"github.com/Kavantix/kantui/internal/column"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Main struct {
	quitting bool
	columns  []column.Model
}

var _ tea.Model = Main{}

func New() Main {
	m := Main{}
	m.columns = []column.Model{
		column.New("TODO"),
		column.New("IN PROGESS"),
		column.New("DONE"),
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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		width := msg.Width / len(m.columns)
		for _, column := range m.columns {
			column.SetSize(width, msg.Height)
		}

	case tea.QuitMsg:
		slog.Info("Quitting")
		m.quitting = true
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "left", "h":
			for i, column := range m.columns {
				if column.Focussed() {
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
				if column.Focussed() {
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

	var cmd tea.Cmd
	for i, column := range m.columns {
		if column.Focussed() {
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
	for _, column := range m.columns {
		columns = append(columns, column.View())
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		columns...,
	)
}
