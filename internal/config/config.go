package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const (
	appDirName            = "wgt"
	defaultConfigFileName = "config.huml"

	DefaultProfileName = "default"
)

var (
	userCacheDirFn  = os.UserCacheDir
	userConfigDirFn = os.UserConfigDir
	userHomeDirFn   = os.UserHomeDir
)

type Config struct {
	DefaultProfile string             `koanf:"default_profile"`
	Profiles       map[string]Profile `koanf:"profiles"`

	// Profile, Server and SSH hold the active profile, set via Use. Top-level
	// server/ssh keys are only read to migrate legacy single-upstream configs.
	Profile string       `koanf:"-"`
	Server  ServerConfig `koanf:"server"`
	SSH     SSHConfig    `koanf:"ssh"`

	Cache CacheConfig `koanf:"cache"`
	UI    UIConfig    `koanf:"ui"`
	Keys  KeysConfig  `koanf:"keys"`
}

type Profile struct {
	Server ServerConfig `koanf:"server"`
	SSH    SSHConfig    `koanf:"ssh"`
}

type ServerConfig struct {
	URL                   string `koanf:"url"`
	Token                 string `koanf:"token"`
	InsecureSkipTLSVerify bool   `koanf:"insecure_skip_tls_verify"`
}

type SSHConfig struct {
	Username  string   `koanf:"username"`
	Host      string   `koanf:"host"`
	Port      int      `koanf:"port"`
	Binary    string   `koanf:"binary"`
	ExtraArgs []string `koanf:"extra_args"`
}

type CacheConfig struct {
	Dir             string        `koanf:"dir"`
	TTL             time.Duration `koanf:"ttl"`
	MaxAge          time.Duration `koanf:"max_age"`
	UseStaleOnError bool          `koanf:"use_stale_on_error"`
}

const (
	ApprovalOpenModeAsk    = "ask"
	ApprovalOpenModeAlways = "always"
	ApprovalOpenModeNever  = "never"

	DefaultApprovalOpenCommand = "xdg-open %s"
)

type UIConfig struct {
	DetailsPane         string `koanf:"details_pane"`
	PreviewLines        int    `koanf:"preview_lines"`
	ApprovalOpenMode    string `koanf:"approval_open_mode"`
	ApprovalOpenCommand string `koanf:"approval_open_command"`
}

type KeysConfig struct {
	Up         []string `koanf:"up"`
	Down       []string `koanf:"down"`
	PageUp     []string `koanf:"page_up"`
	PageDown   []string `koanf:"page_down"`
	Search     []string `koanf:"search"`
	Clear      []string `koanf:"clear"`
	Connect    []string `koanf:"connect"`
	Tunnel     []string `koanf:"tunnel"`
	Rsync      []string `koanf:"rsync"`
	Refresh    []string `koanf:"refresh"`
	EditConfig []string `koanf:"edit_config"`
	Profile    []string `koanf:"profile"`
	Copy       []string `koanf:"copy"`
	Help       []string `koanf:"help"`
	Quit       []string `koanf:"quit"`
}

func Default() Config {
	return Config{
		Profile: DefaultProfileName,
		Server:  ServerConfig{},
		SSH: SSHConfig{
			Port:   2222,
			Binary: "ssh",
		},
		Cache: CacheConfig{
			Dir:             DefaultCacheDir(),
			TTL:             10 * time.Minute,
			MaxAge:          7 * 24 * time.Hour,
			UseStaleOnError: true,
		},
		UI: UIConfig{
			DetailsPane:         "right",
			PreviewLines:        8,
			ApprovalOpenMode:    ApprovalOpenModeAsk,
			ApprovalOpenCommand: DefaultApprovalOpenCommand,
		},
		Keys: KeysConfig{
			Up:         []string{"up", "k"},
			Down:       []string{"down", "j"},
			PageUp:     []string{"pgup", "ctrl+b"},
			PageDown:   []string{"pgdown", "ctrl+f"},
			Search:     []string{"/"},
			Clear:      []string{"esc"},
			Connect:    []string{"enter"},
			Tunnel:     []string{"t"},
			Rsync:      []string{"s"},
			Refresh:    []string{"r"},
			EditConfig: []string{"e"},
			Profile:    []string{"p"},
			Copy:       []string{"c"},
			Help:       []string{"?"},
			Quit:       []string{"q", "ctrl+c"},
		},
	}
}

func (cfg *Config) Use(name string) error {
	p, ok := cfg.Profiles[name]
	if !ok {
		return fmt.Errorf("unknown profile %q", name)
	}
	cfg.Profile, cfg.Server, cfg.SSH = name, p.Server, p.SSH
	return nil
}

func (cfg Config) ProfileNames() []string {
	return slices.Sorted(maps.Keys(cfg.Profiles))
}

func (cfg *Config) RenameActive(name string) {
	if name == cfg.Profile {
		return
	}
	profiles := maps.Clone(cfg.Profiles)
	delete(profiles, cfg.Profile)
	if cfg.DefaultProfile == cfg.Profile {
		cfg.DefaultProfile = name
	}
	cfg.Profiles = profiles
	cfg.Profile = name
}

func (cfg *Config) CommitActive() {
	profiles := maps.Clone(cfg.Profiles)
	if profiles == nil {
		profiles = make(map[string]Profile, 1)
	}
	profiles[cfg.Profile] = Profile{Server: cfg.Server, SSH: cfg.SSH}
	cfg.Profiles = profiles
	if cfg.DefaultProfile == "" {
		cfg.DefaultProfile = cfg.Profile
	}
}

func DefaultCacheDir() string {
	return appPath(userCacheDirFn, ".cache")
}

func ConfigPath() string {
	return appPath(userConfigDirFn, ".config", defaultConfigFileName)
}

func appPath(userDirFn func() (string, error), fallbackRoot string, parts ...string) string {
	if dir, err := userDirFn(); err == nil && strings.TrimSpace(dir) != "" {
		return filepath.Join(append([]string{dir, appDirName}, parts...)...)
	}
	home, _ := userHomeDirFn()
	return filepath.Join(append([]string{home, fallbackRoot, appDirName}, parts...)...)
}
