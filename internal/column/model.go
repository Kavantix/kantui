package column

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	focussed bool
	list     *list.Model
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func New(title string) Model {
	listModel := list.New(
		[]list.Item{
			item{title: "test", desc: "This is a test"},
		},
		list.NewDefaultDelegate(), 0, 0,
	)
	listModel.Title = title
	m := Model{
		list: &listModel,
	}

	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Focus() {
	m.focussed = true
}

func (m *Model) Unfocus() {
	m.focussed = false
}

func (m Model) Focussed() bool {
	return m.focussed
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

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if !m.IsCapturingInput() {
				return m, nil
			}
		}
	}
	newListModel, cmd := m.list.Update(msg)
	m.list = &newListModel
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	borderColor := style.GetBorderTopForeground()
	if m.focussed {
		borderColor = lipgloss.Color("63")
	}
	return style.
		Width(m.list.Width()).
		BorderForeground(borderColor).
		Render(m.list.View())
}
