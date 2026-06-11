package ssh

import (
	"bytes"
	"reflect"
	"testing"

	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func TestBuildApprovalOpenCommand(t *testing.T) {
	bin, args, err := buildApprovalOpenCommand("xdg-open %s", "https://warpgate.example.com/@warpgate#/login/123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Fatalf("buildApprovalOpenCommand: %v", err)
	}
	if bin != "xdg-open" {
		t.Fatalf("expected binary %q, got %q", "xdg-open", bin)
	}
	want := []string{"https://warpgate.example.com/@warpgate#/login/123e4567-e89b-12d3-a456-426614174000"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected args: want %#v got %#v", want, args)
	}
}

func TestBuildApprovalOpenCommandRejectsMissingPlaceholder(t *testing.T) {
	if _, _, err := buildApprovalOpenCommand("xdg-open", "https://warpgate.example.com/@warpgate#/login/123e4567-e89b-12d3-a456-426614174000"); err == nil {
		t.Fatal("expected missing placeholder to fail")
	}
}

func TestApprovalURLDetectorHandlesChunkedOutput(t *testing.T) {
	detector := newApprovalURLDetector()
	chunk1 := "Warpgate authentication\nplease open https://warpgate.example.com/@warpgate#/login/123e4567-"
	chunk2 := "e89b-12d3-a456-426614174000 in your browser\n"

	if got := detector.Feed(chunk1); len(got) != 0 {
		t.Fatalf("expected no URL from first chunk, got %#v", got)
	}
	got := detector.Feed(chunk2)
	want := []string{"https://warpgate.example.com/@warpgate#/login/123e4567-e89b-12d3-a456-426614174000"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected detected URLs: want %#v got %#v", want, got)
	}
}

func TestApprovalURLDetectorAcceptsOptionalSlashBeforeHash(t *testing.T) {
	detector := newApprovalURLDetector()
	got := detector.Feed("https://warpgate.example.com/@warpgate/#/login/123e4567-e89b-12d3-a456-426614174000\n")
	want := []string{"https://warpgate.example.com/@warpgate/#/login/123e4567-e89b-12d3-a456-426614174000"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected detected URLs: want %#v got %#v", want, got)
	}
}

func TestApprovalURLDetectorReportsEachURLOnce(t *testing.T) {
	detector := newApprovalURLDetector()
	chunk := "https://warpgate.example.com/@warpgate#/login/123e4567-e89b-12d3-a456-426614174000\n"
	if got := detector.Feed(chunk); len(got) != 1 {
		t.Fatalf("expected first detection, got %#v", got)
	}
	if got := detector.Feed(chunk); len(got) != 0 {
		t.Fatalf("expected duplicate URL to be ignored, got %#v", got)
	}
}

func TestApprovalURLDetectorWaitsForRemoteEnterPromptMarker(t *testing.T) {
	detector := newApprovalURLDetector()
	detector.Feed("Warpgate authentication\nhttps://warpgate.example.com/@warpgate#/login/123e4567-e89b-12d3-a456-426614174000\n")
	if detector.HasEnterPrompt() {
		t.Fatal("expected enter prompt marker to be absent before the remote prompt arrives")
	}
	detector.Feed("(user:target@warpgate.example.com) Press Enter when done:")
	if !detector.HasEnterPrompt() {
		t.Fatal("expected enter prompt marker to be detected once the remote prompt arrives")
	}
}

func TestHandlePromptInputTreatsEscapeAsNo(t *testing.T) {
	var out bytes.Buffer
	cmd := newConnectExecCommand("test", cfgpkg.Default(), "target")
	cmd.stdout = &out

	done, err := cmd.handlePromptInput("https://warpgate.example.com/@warpgate#/login/123e4567-e89b-12d3-a456-426614174000", []byte{27})
	if err != nil {
		t.Fatalf("handlePromptInput: %v", err)
	}
	if !done {
		t.Fatal("expected escape to dismiss the local prompt")
	}
}

func TestHandlePromptInputTreatsCtrlCAsNo(t *testing.T) {
	var out bytes.Buffer
	cmd := newConnectExecCommand("test", cfgpkg.Default(), "target")
	cmd.stdout = &out

	done, err := cmd.handlePromptInput("https://warpgate.example.com/@warpgate#/login/123e4567-e89b-12d3-a456-426614174000", []byte{3})
	if err != nil {
		t.Fatalf("handlePromptInput: %v", err)
	}
	if !done {
		t.Fatal("expected ctrl-c to dismiss the local prompt")
	}
}

func TestHandlePromptInputTreatsEnterAsNo(t *testing.T) {
	var out bytes.Buffer
	cmd := newConnectExecCommand("test", cfgpkg.Default(), "target")
	cmd.stdout = &out

	done, err := cmd.handlePromptInput("https://warpgate.example.com/@warpgate#/login/123e4567-e89b-12d3-a456-426614174000", []byte{'\r'})
	if err != nil {
		t.Fatalf("handlePromptInput: %v", err)
	}
	if !done {
		t.Fatal("expected enter to dismiss the local prompt")
	}
}
