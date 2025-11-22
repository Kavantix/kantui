package overlay

import (
	"github.com/Kavantix/kantui/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
)

type ModalModel interface {
	tea.Model
	Size() (width, height int)
}

type Sizeable interface {
	SetSize(width, height int) ModalModel
}

type Model struct {
	modals       []ModalModel
	windowWidth  int
	windowHeight int
}

func (m Model) Focused() bool {
	return len(m.modals) > 0
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
	case ModalModel:
		m.modals = append(m.modals, msg)
		return m, nil
	}

	if len(m.modals) > 0 {
		switch msg.(type) {
		case messages.CloseModalMsg:
			m.modals = m.modals[:len(m.modals)-1]
			return m, nil
		}
		i := len(m.modals) - 1
		model, cmd := m.modals[i].Update(msg)
		m.modals[i] = model.(ModalModel)
		return m, cmd
	}

	return m, nil
}

func (m Model) View(app string) string {
	result := app
	for i, modal := range m.modals {
		horizontalMargin := 2
		verticalMargin := 1

		width := m.windowWidth - 2*horizontalMargin
		height := m.windowHeight - 2*verticalMargin
		if sizeable, ok := modal.(Sizeable); ok {
			modal = sizeable.SetSize(width, height)
			m.modals[i] = modal
		}
		width, height = modal.Size()

		result = Place(
			horizontalMargin+((m.windowWidth-2*horizontalMargin)-width)/2,
			verticalMargin+((m.windowHeight-2*verticalMargin)-height)/2,
			modal.View(), DimmBackground(result),
			false,
		)
	}
	return result
}
