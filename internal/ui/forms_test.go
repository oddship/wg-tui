package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func TestParseTransferFlagsSplitsWords(t *testing.T) {
	got := parseTransferFlags("-avz --delete --exclude tmp")
	want := []string{"-avz", "--delete", "--exclude", "tmp"}
	if len(got) != len(want) {
		t.Fatalf("expected %d flags, got %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("flag %d: want %q got %q", i, want[i], got[i])
		}
	}
}

func TestDefaultTransferFlags(t *testing.T) {
	if got := defaultTransferFlags("rsync"); got != "-avz" {
		t.Fatalf("unexpected rsync default flags: %q", got)
	}
	if got := defaultTransferFlags("scp"); got != "-C -p" {
		t.Fatalf("unexpected scp default flags: %q", got)
	}
}

func TestParseTransferToolAliases(t *testing.T) {
	cases := map[string]string{
		"rsync": "rsync",
		"sync":  "rsync",
		"scp":   "scp",
		"copy":  "scp",
	}
	for input, want := range cases {
		got, err := parseTransferTool(input)
		if err != nil {
			t.Fatalf("parseTransferTool(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("expected %q for %q, got %q", want, input, got)
		}
	}
}

func TestParseTransferToolRejectsInvalidValue(t *testing.T) {
	if _, err := parseTransferTool("tar"); err == nil {
		t.Fatal("expected invalid transfer tool to fail")
	}
}

func TestParseRsyncDirectionUploadAliases(t *testing.T) {
	cases := []string{"upload", "up", "push", "host-to-remote", "local-to-remote"}
	for _, input := range cases {
		got, err := parseRsyncDirection(input)
		if err != nil {
			t.Fatalf("parseRsyncDirection(%q): %v", input, err)
		}
		if got != "upload" {
			t.Fatalf("expected upload for %q, got %q", input, got)
		}
	}
}

func TestParseRsyncDirectionDownloadAliases(t *testing.T) {
	cases := []string{"download", "down", "pull", "remote-to-host", "remote-to-local"}
	for _, input := range cases {
		got, err := parseRsyncDirection(input)
		if err != nil {
			t.Fatalf("parseRsyncDirection(%q): %v", input, err)
		}
		if got != "download" {
			t.Fatalf("expected download for %q, got %q", input, got)
		}
	}
}

func TestParseRsyncDirectionRejectsInvalidValue(t *testing.T) {
	if _, err := parseRsyncDirection("sideways"); err == nil {
		t.Fatal("expected invalid rsync direction to fail")
	}
}

func TestParseTunnelPort(t *testing.T) {
	port, err := parseTunnelPort("8000", "remote")
	if err != nil {
		t.Fatalf("parseTunnelPort: %v", err)
	}
	if port != 8000 {
		t.Fatalf("unexpected port: %d", port)
	}
}

func TestParseTunnelPortRejectsInvalidValue(t *testing.T) {
	if _, err := parseTunnelPort("0", "local"); err == nil {
		t.Fatal("expected invalid tunnel port to fail")
	}
}

func TestSyncTunnelFormFieldsMirrorsRemoteWhileLocalUntouched(t *testing.T) {
	m := New("test")
	m.startTunnelForm("svc")
	m.fields[0].input.SetValue("8000")
	m.syncTunnelFormFields(0, "", "8000")

	if got := m.fields[1].input.Value(); got != "8000" {
		t.Fatalf("expected local port to mirror remote, got %q", got)
	}
}

func TestSyncTunnelFormFieldsDoesNotOverwriteTouchedLocal(t *testing.T) {
	m := New("test")
	m.startTunnelForm("svc")
	m.fields[1].input.SetValue("9000")
	m.syncTunnelFormFields(1, "", "9000")
	m.fields[0].input.SetValue("8000")
	m.syncTunnelFormFields(0, "", "8000")

	if got := m.fields[1].input.Value(); got != "9000" {
		t.Fatalf("expected touched local port to stay unchanged, got %q", got)
	}
}

func TestStartTunnelFormKeepsMirroringForRememberedMatchingPorts(t *testing.T) {
	m := New("test")
	m.tunnelLastRemotePort = 8000
	m.tunnelLastLocalPort = 8000
	m.startTunnelForm("svc")
	m.fields[0].input.SetValue("9000")
	m.syncTunnelFormFields(0, "8000", "9000")

	if got := m.fields[1].input.Value(); got != "9000" {
		t.Fatalf("expected remembered matching local port to keep mirroring, got %q", got)
	}
}

func TestStartTunnelFormPreservesRememberedCustomLocalPort(t *testing.T) {
	m := New("test")
	m.tunnelLastRemotePort = 8000
	m.tunnelLastLocalPort = 9000
	m.startTunnelForm("svc")
	m.fields[0].input.SetValue("7000")
	m.syncTunnelFormFields(0, "8000", "7000")

	if got := m.fields[1].input.Value(); got != "9000" {
		t.Fatalf("expected remembered custom local port to stay unchanged, got %q", got)
	}
}

func TestStartRsyncFormUsesSelectFieldsForToolAndDirection(t *testing.T) {
	m := New("test")
	m.startRsyncForm("svc")

	if !m.fields[0].isSelect() {
		t.Fatal("expected tool field to be selectable")
	}
	if !m.fields[1].isSelect() {
		t.Fatal("expected direction field to be selectable")
	}
	if got := m.fields[0].input.Value(); got != "rsync" {
		t.Fatalf("expected default tool rsync, got %q", got)
	}
	if got := m.fields[1].input.Value(); got != "upload" {
		t.Fatalf("expected default direction upload, got %q", got)
	}
	if got := m.fields[2].input.Value(); got != "-avz" {
		t.Fatalf("expected default flags -avz, got %q", got)
	}
}

func TestShiftFieldOptionCyclesSelectValues(t *testing.T) {
	m := New("test")
	m.startRsyncForm("svc")

	m.shiftFieldOption(0, 1)
	if got := m.fields[0].input.Value(); got != "scp" {
		t.Fatalf("expected tool to cycle to scp, got %q", got)
	}
	if got := m.fields[2].input.Value(); got != "-C -p" {
		t.Fatalf("expected flags to switch to scp defaults, got %q", got)
	}

	m.shiftFieldOption(1, 1)
	if got := m.fields[1].input.Value(); got != "download" {
		t.Fatalf("expected direction to cycle to download, got %q", got)
	}

	m.shiftFieldOption(1, 1)
	if got := m.fields[1].input.Value(); got != "upload" {
		t.Fatalf("expected direction to wrap to upload, got %q", got)
	}
}

func TestRememberTransferDraftTracksPathsWhileEditing(t *testing.T) {
	m := New("test")
	m.startRsyncForm("svc")

	m.fields[3].input.SetValue("./dist")
	m.rememberTransferDraft(3)
	m.fields[4].input.SetValue("/srv/app")
	m.rememberTransferDraft(4)

	if got := m.rsyncLastLocalPath; got != "./dist" {
		t.Fatalf("expected remembered local path, got %q", got)
	}
	if got := m.rsyncLastRemotePath; got != "/srv/app" {
		t.Fatalf("expected remembered remote path, got %q", got)
	}
}

func TestStartRsyncFormReusesRememberedPaths(t *testing.T) {
	m := New("test")
	m.rsyncLastLocalPath = "./dist"
	m.rsyncLastRemotePath = "/srv/app"
	m.startRsyncForm("svc")

	if got := m.fields[3].input.Value(); got != "./dist" {
		t.Fatalf("expected remembered local path in form, got %q", got)
	}
	if got := m.fields[4].input.Value(); got != "/srv/app" {
		t.Fatalf("expected remembered remote path in form, got %q", got)
	}
}

func TestStartRsyncFormSetsConditionalHintsForUpload(t *testing.T) {
	m := New("test")
	m.startRsyncForm("svc")

	if got := m.fields[3].hint; got != "Source path on this machine" {
		t.Fatalf("unexpected local upload hint: %q", got)
	}
	if got := m.fields[4].hint; got != "Destination path on the target" {
		t.Fatalf("unexpected remote upload hint: %q", got)
	}
}

func TestShiftFieldOptionUpdatesConditionalHintsForDownload(t *testing.T) {
	m := New("test")
	m.startRsyncForm("svc")

	m.shiftFieldOption(1, 1)

	if got := m.fields[3].hint; got != "Destination path on this machine" {
		t.Fatalf("unexpected local download hint: %q", got)
	}
	if got := m.fields[4].hint; got != "Source path on the target" {
		t.Fatalf("unexpected remote download hint: %q", got)
	}
}

func TestAPITokenHintUsesServerURL(t *testing.T) {
	got := apiTokenHint("https://warpgate.example.com/")
	want := "https://warpgate.example.com/@warpgate/#/profile/api-tokens"
	if got != want {
		t.Fatalf("expected token hint %q, got %q", want, got)
	}
}

func TestAPITokenHintFallsBackForInvalidServerURL(t *testing.T) {
	got := apiTokenHint("not a url")
	want := "Generate this once from the Warpgate web UI"
	if got != want {
		t.Fatalf("expected fallback token hint %q, got %q", want, got)
	}
}

func TestStartFormUsesServerURLForTokenHint(t *testing.T) {
	m := New("test")
	m.cfg = cfgpkg.Default()
	m.cfg.Server.URL = "https://warpgate.example.com/"
	m.startForm(true)

	if got := m.fields[1].hint; got != "https://warpgate.example.com/@warpgate/#/profile/api-tokens" {
		t.Fatalf("expected token hint from server url, got %q", got)
	}
}

func TestEditingServerURLUpdatesTokenHint(t *testing.T) {
	m := New("test")
	m.cfg = cfgpkg.Default()
	m.startForm(true)
	m.fields[0].input.SetValue("")
	m.focusField(0)

	updatedAny, _ := m.updateConfigForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("https://warpgate.example.com/")})
	updated := updatedAny.(Model)

	if got := updated.fields[1].hint; got != "https://warpgate.example.com/@warpgate/#/profile/api-tokens" {
		t.Fatalf("expected token hint to update from edited server url, got %q", got)
	}
}
