package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocCaptureScenarios(t *testing.T) {
	outDir := strings.TrimSpace(os.Getenv("WG_TUI_CAPTURE_OUT"))
	if outDir == "" {
		t.Skip("set WG_TUI_CAPTURE_OUT to generate docs capture artifacts")
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", outDir, err)
	}

	scenarios, err := BuildDocCaptureScenarios()
	if err != nil {
		t.Fatalf("build doc capture scenarios: %v", err)
	}

	for _, scenario := range scenarios {
		if scenario.View != "" {
			writeDocCaptureFile(t, filepath.Join(outDir, scenario.Name+".ansi"), scenario.View)
			continue
		}
		frameDir := filepath.Join(outDir, scenario.Name)
		if err := os.MkdirAll(frameDir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", frameDir, err)
		}
		for i, view := range scenario.Frames {
			name := filepath.Join(frameDir, frameFileName(i+1))
			writeDocCaptureFile(t, name, view)
		}
	}
}

func writeDocCaptureFile(t *testing.T, path, view string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(view), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
