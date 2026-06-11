package ssh

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"unicode"

	"github.com/charmbracelet/x/term"
	"github.com/creack/pty"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
	"golang.org/x/sys/unix"
)

const (
	approvalDetectorTailSize  = 8192
	approvalEnterPromptMarker = "Press Enter when done:"
)

var approvalURLPattern = regexp.MustCompile(`https://[^\s]+/@warpgate/?#/login/[0-9a-fA-F-]{36}`)

type approvalURLDetector struct {
	tail string
	seen map[string]bool
}

type connectExecCommand struct {
	cfgPath       string
	cfg           cfgpkg.Config
	target        string
	stdin         io.Reader
	stdout        io.Writer
	stderr        io.Writer
	status        string
	configUpdated bool
}

func newApprovalURLDetector() *approvalURLDetector {
	return &approvalURLDetector{seen: make(map[string]bool)}
}

func newConnectExecCommand(cfgPath string, cfg cfgpkg.Config, target string) *connectExecCommand {
	return &connectExecCommand{cfgPath: cfgPath, cfg: cfg, target: target}
}

func (d *approvalURLDetector) Feed(chunk string) []string {
	chunk = strings.ReplaceAll(chunk, "\r", "")
	d.tail += chunk
	if len(d.tail) > approvalDetectorTailSize {
		d.tail = d.tail[len(d.tail)-approvalDetectorTailSize:]
	}

	matches := approvalURLPattern.FindAllString(d.tail, -1)
	urls := make([]string, 0, len(matches))
	for _, match := range matches {
		if d.seen[match] {
			continue
		}
		d.seen[match] = true
		urls = append(urls, match)
	}
	return urls
}

func (d *approvalURLDetector) HasEnterPrompt() bool {
	return strings.Contains(d.tail, approvalEnterPromptMarker)
}

func buildApprovalOpenCommand(template, url string) (string, []string, error) {
	template = strings.TrimSpace(template)
	if template == "" {
		return "", nil, fmt.Errorf("approval open command is empty")
	}
	if strings.Count(template, "%s") != 1 {
		return "", nil, fmt.Errorf("approval open command must contain exactly one %%s placeholder")
	}
	parts := strings.Fields(strings.Replace(template, "%s", url, 1))
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("approval open command expanded to an empty command")
	}
	return parts[0], parts[1:], nil
}

func (c *connectExecCommand) SetStdin(r io.Reader)  { c.stdin = r }
func (c *connectExecCommand) SetStdout(w io.Writer) { c.stdout = w }
func (c *connectExecCommand) SetStderr(w io.Writer) { c.stderr = w }

func (c *connectExecCommand) Run() error {
	if c.stdin == nil {
		c.stdin = os.Stdin
	}
	if c.stdout == nil {
		c.stdout = os.Stdout
	}
	if c.stderr == nil {
		c.stderr = os.Stderr
	}

	stdinFile, ok := c.stdin.(*os.File)
	if !ok {
		return fmt.Errorf("interactive ssh requires file-backed stdin")
	}

	bin, args := ShellCommand(c.cfg, c.target)
	cmd := exec.Command(bin, args...)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = ptmx.Close() }()

	if term.IsTerminal(stdinFile.Fd()) {
		state, err := term.MakeRaw(stdinFile.Fd())
		if err != nil {
			return err
		}
		defer func() { _ = term.Restore(stdinFile.Fd(), state) }()

		_ = pty.InheritSize(stdinFile, ptmx)
		resizeSignals := make(chan os.Signal, 1)
		signal.Notify(resizeSignals, unix.SIGWINCH)
		defer signal.Stop(resizeSignals)
		resizeDone := make(chan struct{})
		defer close(resizeDone)
		go func() {
			for {
				select {
				case <-resizeSignals:
					_ = pty.InheritSize(stdinFile, ptmx)
				case <-resizeDone:
					return
				}
			}
		}()
	}

	detector := newApprovalURLDetector()
	promptURL := ""
	pendingPromptURL := ""
	buf := make([]byte, 4096)
	fds := []unix.PollFd{
		{Fd: int32(stdinFile.Fd()), Events: unix.POLLIN},
		{Fd: int32(ptmx.Fd()), Events: unix.POLLIN},
	}

	for {
		_, err := unix.Poll(fds, -1)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return err
		}

		if fds[1].Revents&(unix.POLLIN|unix.POLLHUP|unix.POLLERR) != 0 {
			n, readErr := ptmx.Read(buf)
			if n > 0 {
				chunk := buf[:n]
				if err := writeAll(c.stdout, chunk); err != nil {
					return err
				}
				for _, url := range detector.Feed(string(chunk)) {
					if err := c.handleApprovalURL(url, &promptURL, &pendingPromptURL); err != nil {
						c.writeNotice(fmt.Sprintf("failed to handle approval link: %v", err))
					}
				}
				if promptURL == "" && pendingPromptURL != "" && detector.HasEnterPrompt() {
					promptURL = pendingPromptURL
					pendingPromptURL = ""
					if err := c.showApprovalPrompt(); err != nil {
						return err
					}
				}
			}
			if isPTYEOF(readErr) {
				break
			}
			if readErr != nil {
				return readErr
			}
		}

		if fds[0].Revents&(unix.POLLIN|unix.POLLHUP|unix.POLLERR) != 0 {
			n, readErr := stdinFile.Read(buf)
			if n > 0 {
				input := append([]byte(nil), buf[:n]...)
				if promptURL == "" {
					if err := writeAll(ptmx, input); err != nil {
						return err
					}
				} else {
					done, err := c.handlePromptInput(promptURL, input)
					if err != nil {
						c.writeNotice(fmt.Sprintf("failed to process approval choice: %v", err))
						done = true
					}
					if done {
						promptURL = ""
					}
				}
			}
			if readErr != nil && readErr != io.EOF {
				return readErr
			}
		}
	}

	return cmd.Wait()
}

func (c *connectExecCommand) handleApprovalURL(url string, promptURL, pendingPromptURL *string) error {
	switch c.cfg.UI.ApprovalOpenMode {
	case cfgpkg.ApprovalOpenModeNever:
		return nil
	case cfgpkg.ApprovalOpenModeAlways:
		return c.openApprovalURL(url)
	case cfgpkg.ApprovalOpenModeAsk:
		if *promptURL == "" && *pendingPromptURL == "" {
			*pendingPromptURL = url
		}
		return nil
	default:
		return nil
	}
}

func (c *connectExecCommand) showApprovalPrompt() error {
	_, err := fmt.Fprintf(c.stdout, "\n\rwgt: open approval link in browser? [o]pen once / [a]lways / [n]o: ")
	return err
}

func (c *connectExecCommand) handlePromptInput(promptURL string, input []byte) (bool, error) {
	for _, b := range input {
		r := unicode.ToLower(rune(b))
		switch r {
		case 'o':
			_, _ = fmt.Fprintln(c.stdout, "o")
			return true, c.openApprovalURL(promptURL)
		case 'a':
			_, _ = fmt.Fprintln(c.stdout, "a")
			if err := c.persistApprovalAlways(); err != nil {
				c.writeNotice(fmt.Sprintf("failed to save approval preference: %v", err))
			}
			return true, c.openApprovalURL(promptURL)
		case 'n', '\x1b', '\x03', '\r', '\n':
			_, _ = fmt.Fprintln(c.stdout, "n")
			return true, nil
		case ' ', '\t':
			continue
		default:
			_, _ = fmt.Fprintf(c.stdout, "\n\rwgt: please choose [o], [a], or [n]: ")
		}
	}
	return false, nil
}

func (c *connectExecCommand) persistApprovalAlways() error {
	c.cfg.UI.ApprovalOpenMode = cfgpkg.ApprovalOpenModeAlways
	if err := cfgpkg.Save(c.cfgPath, c.cfg); err != nil {
		return err
	}
	c.configUpdated = true
	c.status = "ssh session ended - saved approval browser preference"
	return nil
}

func (c *connectExecCommand) openApprovalURL(url string) error {
	bin, args, err := buildApprovalOpenCommand(c.cfg.UI.ApprovalOpenCommand, url)
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, args...)
	cmd.Stdout = c.stderr
	cmd.Stderr = c.stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

func (c *connectExecCommand) writeNotice(message string) {
	_, _ = fmt.Fprintf(c.stderr, "\n\rwgt: %s\n", message)
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}
		if n <= 0 {
			return fmt.Errorf("short write")
		}
		data = data[n:]
	}
	return nil
}

func isPTYEOF(err error) bool {
	if err == nil || err == io.EOF {
		return err == io.EOF
	}
	return strings.Contains(strings.ToLower(err.Error()), "input/output error")
}
