package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func ShellCommand(cfg cfgpkg.Config, target string) (string, []string) {
	args := baseArgs(cfg, target)
	return cfg.SSH.Binary, args
}

func TunnelCommand(cfg cfgpkg.Config, target string, remotePort, localPort int) (string, []string) {
	args := []string{
		"-N",
		"-o", "ExitOnForwardFailure=yes",
		"-L", fmt.Sprintf("%d:127.0.0.1:%d", localPort, remotePort),
	}
	args = append(args, baseArgs(cfg, target)...)
	return cfg.SSH.Binary, args
}

func Command(cfg cfgpkg.Config, target string) (string, []string) {
	return ShellCommand(cfg, target)
}

func RsyncCommand(cfg cfgpkg.Config, target, direction string, flags []string, localPath, remotePath string) (string, []string, error) {
	direction = strings.ToLower(strings.TrimSpace(direction))
	if direction != "upload" && direction != "download" {
		return "", nil, fmt.Errorf("invalid rsync direction: %s", direction)
	}
	if strings.TrimSpace(localPath) == "" {
		return "", nil, fmt.Errorf("local path is required")
	}
	if strings.TrimSpace(remotePath) == "" {
		return "", nil, fmt.Errorf("remote path is required")
	}

	remoteShell := shellJoin(rsyncSSHArgs(cfg, target))
	remoteEndpoint := fmt.Sprintf("%s:%s", cfg.SSH.Host, remotePath)

	args := append([]string{}, flags...)
	args = append(args, "-e", remoteShell)
	if direction == "upload" {
		args = append(args, localPath, remoteEndpoint)
	} else {
		args = append(args, remoteEndpoint, localPath)
	}
	return "rsync", args, nil
}

func ScpCommand(cfg cfgpkg.Config, target, direction string, flags []string, localPath, remotePath string) (string, []string, error) {
	direction = strings.ToLower(strings.TrimSpace(direction))
	if direction != "upload" && direction != "download" {
		return "", nil, fmt.Errorf("invalid scp direction: %s", direction)
	}
	if strings.TrimSpace(localPath) == "" {
		return "", nil, fmt.Errorf("local path is required")
	}
	if strings.TrimSpace(remotePath) == "" {
		return "", nil, fmt.Errorf("remote path is required")
	}

	remoteEndpoint := fmt.Sprintf("%s:%s", cfg.SSH.Host, remotePath)
	args := append([]string{}, flags...)
	args = append(args, "-P", strconv.Itoa(cfg.SSH.Port), "-o", fmt.Sprintf("User=%s:%s", cfg.SSH.Username, target))
	args = append(args, cfg.SSH.ExtraArgs...)
	if direction == "upload" {
		args = append(args, localPath, remoteEndpoint)
	} else {
		args = append(args, remoteEndpoint, localPath)
	}
	return "scp", args, nil
}

func ExecCmd(cfg cfgpkg.Config, target string) tea.Cmd {
	bin, args := ShellCommand(cfg, target)
	cmd := exec.Command(bin, args...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return ExecFinishedMsg{Err: err}
	})
}

type ExecFinishedMsg struct{ Err error }

type TunnelProcess struct {
	cfg         cfgpkg.Config
	target      string
	remotePort  int
	localPort   int
	controlPath string
	stderr      bytes.Buffer
}

func PrepareTunnel(cfg cfgpkg.Config, target string, remotePort, localPort int) (*TunnelProcess, error) {
	controlDir := filepath.Join(cfg.Cache.Dir, "ctl")
	if err := os.MkdirAll(controlDir, 0o755); err != nil {
		return nil, err
	}

	controlPath := filepath.Join(controlDir, fmt.Sprintf("%d.sock", time.Now().UnixNano()))
	return &TunnelProcess{
		cfg:         cfg,
		target:      target,
		remotePort:  remotePort,
		localPort:   localPort,
		controlPath: controlPath,
	}, nil
}

func (p *TunnelProcess) StartCmd(fn func(error, string) tea.Msg) tea.Cmd {
	bin, args := interactiveTunnelCommand(p.cfg, p.target, p.remotePort, p.localPort, p.controlPath)
	cmd := exec.Command(bin, args...)
	cmd.Stderr = io.MultiWriter(os.Stderr, &p.stderr)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if fn == nil {
			return nil
		}
		return fn(err, p.StderrSummary())
	})
}

func (p *TunnelProcess) Close() error {
	if p == nil {
		return nil
	}
	if _, err := os.Stat(p.controlPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	bin, args := tunnelControlCommand(p.cfg, p.target, p.controlPath, "exit")
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	_ = os.Remove(p.controlPath)
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		if strings.Contains(msg, "No such file or directory") || strings.Contains(msg, "Control socket connect") {
			return nil
		}
		return fmt.Errorf("close tunnel: %s", msg)
	}
	return nil
}

func (p *TunnelProcess) Check() error {
	if p == nil {
		return fmt.Errorf("tunnel not configured")
	}
	bin, args := tunnelControlCommand(p.cfg, p.target, p.controlPath, "check")
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("check tunnel: %s", msg)
	}
	return nil
}

func (p *TunnelProcess) Cleanup() {
	if p == nil {
		return
	}
	_ = os.Remove(p.controlPath)
}

func (p *TunnelProcess) StderrSummary() string {
	if p == nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(p.stderr.String()), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}
	return ""
}

func interactiveTunnelCommand(cfg cfgpkg.Config, target string, remotePort, localPort int, controlPath string) (string, []string) {
	args := []string{
		"-f",
		"-N",
		"-M",
		"-S", controlPath,
		"-o", "ExitOnForwardFailure=yes",
		"-L", fmt.Sprintf("%d:127.0.0.1:%d", localPort, remotePort),
	}
	args = append(args, baseArgs(cfg, target)...)
	return cfg.SSH.Binary, args
}

func tunnelControlCommand(cfg cfgpkg.Config, target, controlPath, operation string) (string, []string) {
	args := []string{"-S", controlPath, "-O", operation}
	args = append(args, baseArgs(cfg, target)...)
	return cfg.SSH.Binary, args
}

func baseArgs(cfg cfgpkg.Config, target string) []string {
	args := []string{"-p", strconv.Itoa(cfg.SSH.Port), "-l", fmt.Sprintf("%s:%s", cfg.SSH.Username, target)}
	args = append(args, cfg.SSH.ExtraArgs...)
	args = append(args, cfg.SSH.Host)
	return args
}

func rsyncSSHArgs(cfg cfgpkg.Config, target string) []string {
	args := []string{cfg.SSH.Binary, "-p", strconv.Itoa(cfg.SSH.Port), "-l", fmt.Sprintf("%s:%s", cfg.SSH.Username, target)}
	args = append(args, cfg.SSH.ExtraArgs...)
	return args
}

func shellJoin(parts []string) string {
	quoted := make([]string, 0, len(parts))
	for _, part := range parts {
		quoted = append(quoted, shellQuote(part))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if !strings.ContainsAny(value, " \t\n'\"\\$&;|<>*?()[]{}!#~:@") {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
