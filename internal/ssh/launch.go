package ssh

import (
	"fmt"
	"os/exec"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func Command(cfg cfgpkg.Config, target string) (string, []string) {
	args := []string{"-p", strconv.Itoa(cfg.SSH.Port), "-l", fmt.Sprintf("%s:%s", cfg.SSH.Username, target)}
	args = append(args, cfg.SSH.ExtraArgs...)
	args = append(args, cfg.SSH.Host)
	return cfg.SSH.Binary, args
}

func ExecCmd(cfg cfgpkg.Config, target string) tea.Cmd {
	bin, args := Command(cfg, target)
	cmd := exec.Command(bin, args...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return ExecFinishedMsg{Err: err}
	})
}

type ExecFinishedMsg struct{ Err error }
