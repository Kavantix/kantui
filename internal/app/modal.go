package app

import tea "github.com/charmbracelet/bubbletea"

type ModalModel interface {
	tea.Model
	OverlayTitle() string
}

type Sizeable[T any] interface {
	SetSize(width, height int) T
}
