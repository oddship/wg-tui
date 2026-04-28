package config

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Search     key.Binding
	Clear      key.Binding
	Connect    key.Binding
	Refresh    key.Binding
	EditConfig key.Binding
	Copy       key.Binding
	Help       key.Binding
	Quit       key.Binding
}

func NewKeyMap(cfg KeysConfig) KeyMap {
	return KeyMap{
		Up:         bind(cfg.Up, "up", "up"),
		Down:       bind(cfg.Down, "down", "down"),
		Search:     bind(cfg.Search, "/", "search"),
		Clear:      bind(cfg.Clear, "esc", "clear"),
		Connect:    bind(cfg.Connect, "enter", "connect"),
		Refresh:    bind(cfg.Refresh, "r", "refresh"),
		EditConfig: bind(cfg.EditConfig, "e", "config"),
		Copy:       bind(cfg.Copy, "c", "copy ssh"),
		Help:       bind(cfg.Help, "?", "help"),
		Quit:       bind(cfg.Quit, "q", "quit"),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Search, k.Connect, k.Refresh, k.EditConfig, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Search, k.Clear, k.Connect}, {k.Refresh, k.EditConfig, k.Copy, k.Help, k.Quit}}
}

func bind(keys []string, label, desc string) key.Binding {
	if len(keys) == 0 {
		keys = []string{label}
	}
	return key.NewBinding(key.WithKeys(keys...), key.WithHelp(keys[0], desc))
}
