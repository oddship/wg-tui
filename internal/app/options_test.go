package app

import (
	"bytes"
	"strings"
	"testing"

	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func TestParseOptionsDefaultsToConfigPath(t *testing.T) {
	opts, err := ParseOptions(nil)
	if err != nil {
		t.Fatalf("ParseOptions: %v", err)
	}
	if opts.ConfigPath != cfgpkg.ConfigPath() {
		t.Fatalf("expected default config path %q, got %q", cfgpkg.ConfigPath(), opts.ConfigPath)
	}
	if opts.CachePath != "" {
		t.Fatalf("expected empty cache override, got %q", opts.CachePath)
	}
}

func TestParseOptionsAcceptsOverrides(t *testing.T) {
	opts, err := ParseOptions([]string{"--config", "/tmp/custom.huml", "--cache-path", "/tmp/cache"})
	if err != nil {
		t.Fatalf("ParseOptions: %v", err)
	}
	if opts.ConfigPath != "/tmp/custom.huml" {
		t.Fatalf("expected config path override, got %q", opts.ConfigPath)
	}
	if opts.CachePath != "/tmp/cache" {
		t.Fatalf("expected cache path override, got %q", opts.CachePath)
	}
}

func TestParseOptionsRecognizesHelp(t *testing.T) {
	opts, err := ParseOptions([]string{"--help"})
	if err != nil {
		t.Fatalf("ParseOptions: %v", err)
	}
	if !opts.ShowHelp {
		t.Fatal("expected help flag to be recognized")
	}
}

func TestParseOptionsRejectsUnexpectedArgs(t *testing.T) {
	if _, err := ParseOptions([]string{"extra"}); err == nil {
		t.Fatal("expected unexpected positional args to fail")
	}
}

func TestRunHelpPrintsUsage(t *testing.T) {
	var out bytes.Buffer
	if err := Run([]string{"--help"}, &out); err != nil {
		t.Fatalf("Run(help): %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "Usage: wgt [options]") {
		t.Fatalf("expected usage header, got %q", text)
	}
	if !strings.Contains(text, "--cache-path <dir>") {
		t.Fatalf("expected cache-path flag in help, got %q", text)
	}
}
