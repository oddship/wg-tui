package config

import (
	"fmt"
	"strings"
)

func validate(cfg Config) error {
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

func normalizeKey(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}
