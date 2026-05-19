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

func TestRsyncCommandUpload(t *testing.T) {
	cfg := cfgpkg.Default()
	cfg.SSH.Username = "alice@example.com"
	cfg.SSH.Host = "warpgate.example.com"
	cfg.SSH.Port = 2222
	cfg.SSH.ExtraArgs = []string{"-o", "ServerAliveInterval=30"}

	bin, args, err := RsyncCommand(cfg, "prod-api", "upload", []string{"-avz"}, "./build/", "/srv/app/")
	if err != nil {
		t.Fatalf("RsyncCommand(upload): %v", err)
	}
	if bin != "rsync" {
		t.Fatalf("unexpected binary: %s", bin)
	}

	want := []string{
		"-avz",
		"-e", "ssh -p 2222 -l 'alice@example.com:prod-api' -o ServerAliveInterval=30",
		"./build/",
		"warpgate.example.com:/srv/app/",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected args:\nwant: %#v\n got: %#v", want, args)
	}
}

func TestRsyncCommandDownload(t *testing.T) {
	cfg := cfgpkg.Default()
	cfg.SSH.Username = "alice@example.com"
	cfg.SSH.Host = "warpgate.example.com"
	cfg.SSH.Port = 2222

	bin, args, err := RsyncCommand(cfg, "prod-api", "download", []string{"-avz"}, "./logs/", "/var/log/app/")
	if err != nil {
		t.Fatalf("RsyncCommand(download): %v", err)
	}
	if bin != "rsync" {
		t.Fatalf("unexpected binary: %s", bin)
	}

	want := []string{
		"-avz",
		"-e", "ssh -p 2222 -l 'alice@example.com:prod-api'",
		"warpgate.example.com:/var/log/app/",
		"./logs/",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected args:\nwant: %#v\n got: %#v", want, args)
	}
}

func TestRsyncCommandRejectsInvalidDirection(t *testing.T) {
	cfg := cfgpkg.Default()
	cfg.SSH.Username = "alice@example.com"
	cfg.SSH.Host = "warpgate.example.com"
	cfg.SSH.Port = 2222

	if _, _, err := RsyncCommand(cfg, "prod-api", "sideways", []string{"-avz"}, "./a", "/b"); err == nil {
		t.Fatal("expected invalid rsync direction to fail")
	}
}

func TestScpCommandUpload(t *testing.T) {
	cfg := cfgpkg.Default()
	cfg.SSH.Username = "alice@example.com"
	cfg.SSH.Host = "warpgate.example.com"
	cfg.SSH.Port = 2222
	cfg.SSH.ExtraArgs = []string{"-o", "ServerAliveInterval=30"}

	bin, args, err := ScpCommand(cfg, "prod-api", "upload", []string{"-C", "-p"}, "./build/app", "/srv/app/")
	if err != nil {
		t.Fatalf("ScpCommand(upload): %v", err)
	}
	if bin != "scp" {
		t.Fatalf("unexpected binary: %s", bin)
	}

	want := []string{
		"-C",
		"-p",
		"-P", "2222",
		"-o", "User=alice@example.com:prod-api",
		"-o", "ServerAliveInterval=30",
		"./build/app",
		"warpgate.example.com:/srv/app/",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected args:\nwant: %#v\n got: %#v", want, args)
	}
}

func TestScpCommandDownload(t *testing.T) {
	cfg := cfgpkg.Default()
	cfg.SSH.Username = "alice@example.com"
	cfg.SSH.Host = "warpgate.example.com"
	cfg.SSH.Port = 2222

	bin, args, err := ScpCommand(cfg, "prod-api", "download", []string{"-C", "-p"}, "./logs/", "/var/log/app.log")
	if err != nil {
		t.Fatalf("ScpCommand(download): %v", err)
	}
	if bin != "scp" {
		t.Fatalf("unexpected binary: %s", bin)
	}

	want := []string{
		"-C",
		"-p",
		"-P", "2222",
		"-o", "User=alice@example.com:prod-api",
		"warpgate.example.com:/var/log/app.log",
		"./logs/",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected args:\nwant: %#v\n got: %#v", want, args)
	}
}

func TestScpCommandRejectsInvalidDirection(t *testing.T) {
	cfg := cfgpkg.Default()
	cfg.SSH.Username = "alice@example.com"
	cfg.SSH.Host = "warpgate.example.com"
	cfg.SSH.Port = 2222

	if _, _, err := ScpCommand(cfg, "prod-api", "sideways", []string{"-C", "-p"}, "./a", "/b"); err == nil {
		t.Fatal("expected invalid scp direction to fail")
	}
}
