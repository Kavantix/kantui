package overlay

import tea "github.com/charmbracelet/bubbletea"

type ModalModel interface {
	tea.Model
	OverlayTitle() string
	Size() (width, height int)
}

type Sizeable interface {
	SetSize(width, height int) tea.Model
}
