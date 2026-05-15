package ui

import "testing"

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
