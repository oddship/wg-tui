package config

import "testing"

func TestValidateAllowsDefaultKeymap(t *testing.T) {
	if err := validate(Default()); err != nil {
		t.Fatalf("validate(default): %v", err)
	}
}

func TestValidateRejectsConflictingBindings(t *testing.T) {
	cfg := Default()
	cfg.Keys.EditConfig = []string{"e"}
	cfg.Keys.Search = []string{"e"}

	err := validate(cfg)
	if err == nil {
		t.Fatal("expected conflicting keybindings to fail validation")
	}
}

func TestValidateRejectsEmptyBinding(t *testing.T) {
	cfg := Default()
	cfg.Keys.Copy = []string{""}

	err := validate(cfg)
	if err == nil {
		t.Fatal("expected empty keybinding to fail validation")
	}
}
