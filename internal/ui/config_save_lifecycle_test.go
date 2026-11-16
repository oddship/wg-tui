package ui

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func configEditorModel(t *testing.T) Model {
	t.Helper()
	m := multiProfileModel(t)
	m.cfgPath = filepath.Join(t.TempDir(), "config.huml")
	m.cacheDirOverride = t.TempDir()
	if err := cfgpkg.Save(m.cfgPath, m.cfg); err != nil {
		t.Fatal(err)
	}
	m.openConfigEditor()
	return m
}

func TestConfigSaveBlocksNavigationEditsAndDuplicateSubmissions(t *testing.T) {
	m := configEditorModel(t)
	requireField(t, m.fields, "ui.target_contrast").input.SetValue("subtle")
	updated, save := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(Model)
	if save == nil || !m.configSaving {
		t.Fatal("save did not start")
	}
	generation := m.configSaveGeneration
	name := requireField(t, m.fields, "profile.name").input.Value()
	for _, key := range []tea.KeyMsg{{Type: tea.KeyEsc}, {Type: tea.KeyEnter}, {Type: tea.KeyCtrlS}, {Type: tea.KeyTab}, {Type: tea.KeyRunes, Runes: []rune("p")}} {
		updated, cmd := m.Update(key)
		m = updated.(Model)
		if cmd != nil || m.mode != modeConfig || m.cfg.Profile != "prod" || m.configSaveGeneration != generation {
			t.Fatal("navigation or duplicate save allowed during persistence")
		}
	}
	if requireField(t, m.fields, "profile.name").input.Value() != name {
		t.Fatal("input changed during persistence")
	}
	msg := save().(configSavedMsg)
	if msg.err != nil {
		t.Fatal(msg.err)
	}
	updated, _ = m.Update(msg)
	m = updated.(Model)
	if m.configSaving || m.mode != modeBrowse {
		t.Fatal("successful save did not unlock the UI")
	}
	saved, err := cfgpkg.Load(m.cfgPath)
	if err != nil || saved.UI.TargetContrast != "subtle" {
		t.Fatalf("save not persisted: %v", err)
	}
	// Completion replays after the user has moved on must have no effect.
	updated, load := m.switchProfile("staging")
	m = updated.(Model)
	updated, _ = m.Update(load())
	m = updated.(Model)
	updated, cmd := m.Update(msg)
	m = updated.(Model)
	if cmd != nil || m.cfg.Profile != "staging" {
		t.Fatal("stale result switched the upstream")
	}
}

func TestConfigSaveRejectsOldGenerationDuringNewSave(t *testing.T) {
	m := configEditorModel(t)
	updated, save := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(Model)
	stale := configSavedMsg{generation: m.configSaveGeneration - 1, err: fmt.Errorf("old error")}
	updated, _ = m.Update(stale)
	m = updated.(Model)
	if !m.configSaving || strings.Contains(m.formStatus, "old error") {
		t.Fatal("old completion changed active save")
	}
	updated, _ = m.Update(save())
	m = updated.(Model)
	m.openConfigEditor()
	updated, second := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(Model)
	if second == nil || m.configSaveGeneration != 2 {
		t.Fatal("new save did not get a new generation")
	}
}

func TestConfigSaveFailureUnlocksRetryAndPreservesConfig(t *testing.T) {
	m := configEditorModel(t)
	original, err := os.ReadFile(m.cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	// Force an offline rename's backup to fail without changing source files.
	if err := os.WriteFile(filepath.Join(m.cacheDirOverride, ".backups"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	requireField(t, m.fields, "profile.name").input.SetValue("production")
	updated, save := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(Model)
	updated, _ = m.Update(save())
	m = updated.(Model)
	if m.configSaving || m.mode != modeConfig || !strings.Contains(m.formStatus, "rename stopped") {
		t.Fatalf("failure did not unlock editor: %s", m.formStatus)
	}
	after, err := os.ReadFile(m.cfgPath)
	if err != nil || string(original) != string(after) {
		t.Fatal("failed save changed config")
	}
	if err := os.Remove(filepath.Join(m.cacheDirOverride, ".backups")); err != nil {
		t.Fatal(err)
	}
	updated, retry := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(Model)
	if retry == nil {
		t.Fatal("retry blocked")
	}
	updated, _ = m.Update(retry())
	m = updated.(Model)
	if m.configSaving || m.cfg.Profile != "production" {
		t.Fatalf("retry failed: %s", m.formStatus)
	}
}

func TestCtrlCWaitsForConfigSaveBeforeQuitting(t *testing.T) {
	m := configEditorModel(t)
	requireField(t, m.fields, "ui.target_contrast").input.SetValue("subtle")
	updated, save := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(Model)
	updated, quit := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(Model)
	if quit != nil || !m.configSaving || !m.quitAfterConfigSave {
		t.Fatal("quit interrupted save")
	}
	updated, quit = m.Update(save())
	m = updated.(Model)
	if quit == nil || m.configSaving {
		t.Fatal("quit was not delivered after save")
	}
	if _, ok := quit().(tea.QuitMsg); !ok {
		t.Fatal("expected quit")
	}
	saved, err := cfgpkg.Load(m.cfgPath)
	if err != nil || saved.UI.TargetContrast != "subtle" {
		t.Fatal("quit occurred without persisted config")
	}
}

func TestValidationFailureUnlocksConfigEditor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "unavailable", 503) }))
	defer server.Close()
	m := configEditorModel(t)
	requireField(t, m.fields, "server.url").input.SetValue(server.URL)
	updated, save := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(Model)
	updated, _ = m.Update(save())
	m = updated.(Model)
	if m.configSaving || m.mode != modeConfig || !strings.Contains(m.formStatus, "validation failed") {
		t.Fatalf("validation failure left the editor locked: %s", m.formStatus)
	}
}
