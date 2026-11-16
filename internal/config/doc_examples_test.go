package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentedConfigExamplesLoad(t *testing.T) {
	for _, file := range []string{"README.md", "SPEC.md", "docs/02-reference/01-configuration.md"} {
		t.Run(file, func(t *testing.T) {
			body, err := os.ReadFile(filepath.Join("../..", file))
			if err != nil {
				t.Fatal(err)
			}
			_, rest, ok := strings.Cut(string(body), "```huml\n")
			if !ok {
				t.Fatal("missing HUML example")
			}
			example, _, ok := strings.Cut(rest, "```")
			if !ok {
				t.Fatal("unclosed example")
			}
			path := filepath.Join(t.TempDir(), "config.huml")
			if err := os.WriteFile(path, []byte(example), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path)
			if err != nil {
				t.Fatalf("example does not load: %v", err)
			}
			for _, name := range cfg.ProfileNames() {
				if err := cfg.Use(name); err != nil {
					t.Fatal(err)
				}
				if cfg.SSH.Host == "" {
					t.Fatalf("documented profile %q has no SSH hostname", name)
				}
			}
			if cfg.UI.DetailsPane != Default().UI.DetailsPane || cfg.UI.TargetFields != Default().UI.TargetFields || cfg.UI.TargetContrast != Default().UI.TargetContrast {
				t.Fatalf("documented defaults differ from app: %+v", cfg.UI)
			}
		})
	}
}
