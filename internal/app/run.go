package app

import (
	tea "github.com/charmbracelet/bubbletea"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
	"github.com/oddship/wg-tui/internal/ui"
)

func Run() error {
	p := tea.NewProgram(ui.New(cfgpkg.ConfigPath()), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
