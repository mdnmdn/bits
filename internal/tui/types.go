package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/mdnmdn/bits/internal/model"
)

type Msg tea.Msg

type KeyMsg tea.KeyMsg

type WindowSizeMsg tea.WindowSizeMsg

type SectionModel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

type SectionWithProvider interface {
	SectionModel
	ProviderID() string
	Market() model.MarketType
	Cursor() int
	Loading() bool
	Error() error
}
