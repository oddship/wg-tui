package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/huml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

func Load(path string) (Config, error) {
	cfg := Default()
	k := koanf.New(".")
	if err := k.Load(file.Provider(path), huml.Parser()); err != nil {
		return Config{}, err
	}
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, err
	}
	applyDerived(&cfg)
	return cfg, nil
}

func Save(path string, cfg Config) error {
	applyDerived(&cfg)
	b, err := huml.Parser().Marshal(cfg.toMap())
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func (cfg Config) toMap() map[string]any {
	return map[string]any{
		"server": map[string]any{
			"url":                      cfg.Server.URL,
			"token":                    cfg.Server.Token,
			"insecure_skip_tls_verify": cfg.Server.InsecureSkipTLSVerify,
		},
		"ssh": map[string]any{
			"username":   cfg.SSH.Username,
			"host":       cfg.SSH.Host,
			"port":       cfg.SSH.Port,
			"binary":     cfg.SSH.Binary,
			"extra_args": cfg.SSH.ExtraArgs,
		},
		"cache": map[string]any{
			"dir":                cfg.Cache.Dir,
			"ttl":                cfg.Cache.TTL.String(),
			"max_age":            cfg.Cache.MaxAge.String(),
			"use_stale_on_error": cfg.Cache.UseStaleOnError,
		},
		"ui": map[string]any{
			"details_pane":  cfg.UI.DetailsPane,
			"preview_lines": cfg.UI.PreviewLines,
		},
		"keys": cfg.Keys.toMap(),
	}
}

func (k KeysConfig) toMap() map[string]any {
	return map[string]any{
		"up":          k.Up,
		"down":        k.Down,
		"search":      k.Search,
		"clear":       k.Clear,
		"connect":     k.Connect,
		"refresh":     k.Refresh,
		"edit_config": k.EditConfig,
		"copy":        k.Copy,
		"help":        k.Help,
		"quit":        k.Quit,
	}
}

func applyDerived(cfg *Config) {
	def := Default()
	if cfg.SSH.Binary == "" {
		cfg.SSH.Binary = def.SSH.Binary
	}
	if cfg.SSH.Port == 0 {
		cfg.SSH.Port = def.SSH.Port
	}
	if cfg.Cache.Dir == "" {
		cfg.Cache.Dir = def.Cache.Dir
	}
	if cfg.Cache.TTL == 0 {
		cfg.Cache.TTL = def.Cache.TTL
	}
	if cfg.Cache.MaxAge == 0 {
		cfg.Cache.MaxAge = def.Cache.MaxAge
	}
	if cfg.UI.DetailsPane == "" {
		cfg.UI.DetailsPane = def.UI.DetailsPane
	}
	if cfg.UI.PreviewLines == 0 {
		cfg.UI.PreviewLines = def.UI.PreviewLines
	}
	mergeKeys(&cfg.Keys, def.Keys)
}

func mergeKeys(k *KeysConfig, d KeysConfig) {
	if len(k.Up) == 0 {
		k.Up = d.Up
	}
	if len(k.Down) == 0 {
		k.Down = d.Down
	}
	if len(k.Search) == 0 {
		k.Search = d.Search
	}
	if len(k.Clear) == 0 {
		k.Clear = d.Clear
	}
	if len(k.Connect) == 0 {
		k.Connect = d.Connect
	}
	if len(k.Refresh) == 0 {
		k.Refresh = d.Refresh
	}
	if len(k.EditConfig) == 0 {
		k.EditConfig = d.EditConfig
	}
	if len(k.Copy) == 0 {
		k.Copy = d.Copy
	}
	if len(k.Help) == 0 {
		k.Help = d.Help
	}
	if len(k.Quit) == 0 {
		k.Quit = d.Quit
	}
}
