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
	if !k.Exists("keys.tunnel") {
		cfg.Keys.Tunnel = nil
	}
	if !k.Exists("keys.rsync") {
		cfg.Keys.Rsync = nil
	}
	if !k.Exists("keys.profile") {
		cfg.Keys.Profile = nil
	}
	if len(cfg.Profiles) == 0 && k.Exists("server") {
		cfg.Profiles = map[string]Profile{DefaultProfileName: {Server: cfg.Server, SSH: cfg.SSH}}
	}
	if cfg.DefaultProfile == "" && len(cfg.Profiles) > 0 {
		cfg.DefaultProfile = cfg.ProfileNames()[0]
	}
	applyDerived(&cfg)
	if err := validateProfiles(cfg); err != nil {
		return Config{}, err
	}
	if err := validate(cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.Use(cfg.DefaultProfile); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	applyDerived(&cfg)
	if err := validateProfiles(cfg); err != nil {
		return err
	}
	if err := validate(cfg); err != nil {
		return err
	}
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
	profiles := make(map[string]any, len(cfg.Profiles))
	for name, p := range cfg.Profiles {
		profiles[name] = p.toMap()
	}
	return map[string]any{
		"default_profile": cfg.DefaultProfile,
		"profiles":        profiles,
		"cache": map[string]any{
			"dir":                cfg.Cache.Dir,
			"ttl":                cfg.Cache.TTL.String(),
			"max_age":            cfg.Cache.MaxAge.String(),
			"use_stale_on_error": cfg.Cache.UseStaleOnError,
		},
		"ui": map[string]any{
			"details_pane":          cfg.UI.DetailsPane,
			"preview_lines":         cfg.UI.PreviewLines,
			"approval_open_mode":    cfg.UI.ApprovalOpenMode,
			"approval_open_command": cfg.UI.ApprovalOpenCommand,
		},
		"keys": cfg.Keys.toMap(),
	}
}

func (p Profile) toMap() map[string]any {
	return map[string]any{
		"server": map[string]any{
			"url":                      p.Server.URL,
			"token":                    p.Server.Token,
			"insecure_skip_tls_verify": p.Server.InsecureSkipTLSVerify,
		},
		"ssh": map[string]any{
			"username":   p.SSH.Username,
			"host":       p.SSH.Host,
			"port":       p.SSH.Port,
			"binary":     p.SSH.Binary,
			"extra_args": p.SSH.ExtraArgs,
		},
	}
}

func (k KeysConfig) toMap() map[string]any {
	return map[string]any{
		"up":          k.Up,
		"down":        k.Down,
		"page_up":     k.PageUp,
		"page_down":   k.PageDown,
		"search":      k.Search,
		"clear":       k.Clear,
		"connect":     k.Connect,
		"tunnel":      k.Tunnel,
		"rsync":       k.Rsync,
		"refresh":     k.Refresh,
		"edit_config": k.EditConfig,
		"profile":     k.Profile,
		"copy":        k.Copy,
		"help":        k.Help,
		"quit":        k.Quit,
	}
}

func applyDerived(cfg *Config) {
	def := Default()
	applySSHDefaults(&cfg.SSH, def.SSH)
	for name, p := range cfg.Profiles {
		applySSHDefaults(&p.SSH, def.SSH)
		cfg.Profiles[name] = p
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
	if cfg.UI.ApprovalOpenMode == "" {
		cfg.UI.ApprovalOpenMode = def.UI.ApprovalOpenMode
	}
	if cfg.UI.ApprovalOpenCommand == "" {
		cfg.UI.ApprovalOpenCommand = def.UI.ApprovalOpenCommand
	}
	mergeKeys(&cfg.Keys, def.Keys)
}

func applySSHDefaults(s *SSHConfig, def SSHConfig) {
	if s.Binary == "" {
		s.Binary = def.Binary
	}
	if s.Port == 0 {
		s.Port = def.Port
	}
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
