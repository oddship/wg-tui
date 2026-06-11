package config

import (
	"fmt"
	"strings"
)

func validate(cfg Config) error {
	if err := validateApprovalOpenMode(cfg.UI.ApprovalOpenMode); err != nil {
		return err
	}
	if err := validateApprovalOpenCommand(cfg.UI.ApprovalOpenMode, cfg.UI.ApprovalOpenCommand); err != nil {
		return err
	}
	type binding struct {
		name string
		keys []string
	}

	bindings := []binding{
		{name: "up", keys: cfg.Keys.Up},
		{name: "down", keys: cfg.Keys.Down},
		{name: "search", keys: cfg.Keys.Search},
		{name: "clear", keys: cfg.Keys.Clear},
		{name: "connect", keys: cfg.Keys.Connect},
		{name: "tunnel", keys: cfg.Keys.Tunnel},
		{name: "rsync", keys: cfg.Keys.Rsync},
		{name: "refresh", keys: cfg.Keys.Refresh},
		{name: "edit_config", keys: cfg.Keys.EditConfig},
		{name: "copy", keys: cfg.Keys.Copy},
		{name: "help", keys: cfg.Keys.Help},
		{name: "quit", keys: cfg.Keys.Quit},
	}

	seen := make(map[string]string)
	for _, binding := range bindings {
		for _, rawKey := range binding.keys {
			key := normalizeKey(rawKey)
			if key == "" {
				return fmt.Errorf("invalid empty keybinding for %s", binding.name)
			}
			if prev, ok := seen[key]; ok && prev != binding.name {
				return fmt.Errorf("keybinding conflict: %q is assigned to both %s and %s", key, prev, binding.name)
			}
			seen[key] = binding.name
		}
	}

	return nil
}

func validateApprovalOpenMode(mode string) error {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case ApprovalOpenModeAsk, ApprovalOpenModeAlways, ApprovalOpenModeNever:
		return nil
	default:
		return fmt.Errorf("invalid approval_open_mode: %q", mode)
	}
}

func validateApprovalOpenCommand(mode, command string) error {
	if strings.EqualFold(strings.TrimSpace(mode), ApprovalOpenModeNever) {
		return nil
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("approval_open_command is required unless approval_open_mode is never")
	}
	if strings.Count(command, "%s") != 1 {
		return fmt.Errorf("approval_open_command must contain exactly one %%s placeholder")
	}
	return nil
}

func normalizeKey(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}
