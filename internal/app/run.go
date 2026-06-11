package app

import (
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oddship/wg-tui/internal/ui"
)

func Run(args []string, stdout io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}

	opts, err := ParseOptions(args)
	if err != nil {
		return err
	}
	if opts.ShowHelp {
		_, err := io.WriteString(stdout, Usage())
		return err
	}

	p := tea.NewProgram(ui.New(opts.ConfigPath, ui.WithCacheDirOverride(opts.CachePath)), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
