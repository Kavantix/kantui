package confirm

import (
	"github.com/Kavantix/kantui/internal/messages"
	"github.com/Kavantix/kantui/internal/overlay"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	question  string
	onConfirm tea.Cmd
}

// assert
var _ overlay.ModalModel = Model{}

func Show(question string, onConfirm tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		return Model{
			question:  question,
			onConfirm: onConfirm,
		}
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			return m, tea.Batch(m.onConfirm, messages.CloseModal)
		case "n", "esc":
			return m, messages.CloseModal
		}
	}
	return m, nil
}

func (m Model) OverlayTitle() string {
	return "Confirm"
}

// Size implements overlay.ModalModel.
func (m Model) Size() (width int, height int) {
	styleWidth, styleHeight := confirmStyle.GetFrameSize()
	content := m.View()
	return lipgloss.Width(content) + styleWidth, lipgloss.Height(content) + styleHeight
}

var confirmStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(1, 2)

func (m Model) View() string {
	return confirmStyle.Render(m.question + " (y/n)")
}
