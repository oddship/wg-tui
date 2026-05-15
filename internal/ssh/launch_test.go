package ssh

import (
	"reflect"
	"testing"

	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func TestShellCommand(t *testing.T) {
	cfg := cfgpkg.Default()
	cfg.SSH.Username = "alice@example.com"
	cfg.SSH.Host = "warpgate.example.com"
	cfg.SSH.Port = 2222
	cfg.SSH.Binary = "ssh"
	cfg.SSH.ExtraArgs = []string{"-o", "ServerAliveInterval=30"}

	bin, args := ShellCommand(cfg, "prod-api")
	if bin != "ssh" {
		t.Fatalf("unexpected binary: %s", bin)
	}

	want := []string{
		"-p", "2222",
		"-l", "alice@example.com:prod-api",
		"-o", "ServerAliveInterval=30",
		"warpgate.example.com",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected args:\nwant: %#v\n got: %#v", want, args)
	}
}

func TestTunnelCommand(t *testing.T) {
	cfg := cfgpkg.Default()
	cfg.SSH.Username = "alice@example.com"
	cfg.SSH.Host = "warpgate.example.com"
	cfg.SSH.Port = 2222
	cfg.SSH.Binary = "ssh"

	bin, args := TunnelCommand(cfg, "svc-admin", 8000, 9000)
	if bin != "ssh" {
		t.Fatalf("unexpected binary: %s", bin)
	}

	want := []string{
		"-N",
		"-o", "ExitOnForwardFailure=yes",
		"-L", "9000:127.0.0.1:8000",
		"-p", "2222",
		"-l", "alice@example.com:svc-admin",
		"warpgate.example.com",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected args:\nwant: %#v\n got: %#v", want, args)
	}
}
