package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oddship/wg-tui/internal/ui"
)

func main() {
	var scenario string
	var listScenarios bool
	var tapeDir string
	var assetDir string
	var scratchDir string
	var binaryPath string

	fs := flag.NewFlagSet("wgt-docs", flag.ExitOnError)
	fs.StringVar(&scenario, "scenario", "browse-overview", "docs capture scenario")
	fs.BoolVar(&listScenarios, "list-scenarios", false, "list available docs scenarios")
	fs.StringVar(&tapeDir, "write-vhs-tapes", "", "write VHS tape files to this directory")
	fs.StringVar(&assetDir, "asset-dir", "docs/assets", "asset directory path used inside generated VHS tapes")
	fs.StringVar(&scratchDir, "scratch-dir", "workspace/scratch/wg-tui-doc-capture", "scratch directory path used inside generated VHS tapes")
	fs.StringVar(&binaryPath, "binary", "", "path to the docs helper binary used inside generated VHS tapes")
	fs.Parse(os.Args[1:])

	if listScenarios {
		fmt.Println(strings.Join(ui.DocCaptureScenarioNames(), "\n"))
		return
	}

	if strings.TrimSpace(tapeDir) != "" {
		if strings.TrimSpace(binaryPath) == "" {
			fmt.Fprintln(os.Stderr, "--binary is required with --write-vhs-tapes")
			os.Exit(1)
		}
		if err := ui.WriteDocCaptureVHSTapes(tapeDir, binaryPath, assetDir, scratchDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	model, err := ui.NewDocCaptureScenarioModel(scenario)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
