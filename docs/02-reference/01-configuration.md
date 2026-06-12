# Configuration

`wgt` stores config in HUML.

## Default locations

By default, `wgt` uses your platform config directory. Cache and local state go in your platform cache directory.

On Linux, that is usually:

- config: `~/.config/wgt/config.huml`
- cache and local state: `~/.cache/wgt/`

The cache directory stores:

- cached Warpgate target snapshots
- recent targets
- remembered transfer form state

## CLI overrides

`wgt` also supports a few small CLI flags:

```bash
wgt --help
wgt --config /path/to/config.huml
wgt --cache-path /path/to/cache-dir
```

`--cache-path` only affects the current run. It does not rewrite `cache.dir` in your saved config.

## Example config

You only need to set `cache.dir` if you want to override the platform default.

```huml
%HUML v0.2.0

server:
  url: "https://warpgate.example.com"
  token: "<warpgate-api-token>"
  insecure_skip_tls_verify: false

ssh:
  username: "user@example.com"
  host: "warpgate.example.com"
  port: 2222
  binary: "ssh"
  extra_args::
    - "-o"
    - "ServerAliveInterval=30"

cache:
  dir: "/path/to/cache/wgt"
  ttl: "10m"
  max_age: "168h"
  use_stale_on_error: true

ui:
  details_pane: "right"
  preview_lines: 8
  approval_open_mode: "ask"
  approval_open_command: "xdg-open %s"

keys:
  up::
    - "up"
    - "k"
  down::
    - "down"
    - "j"
  search::
    - "/"
  clear::
    - "esc"
  connect::
    - "enter"
  tunnel::
    - "t"
  rsync::
    - "s"
  refresh::
    - "r"
  edit_config::
    - "e"
  copy::
    - "c"
  help::
    - "?"
  quit::
    - "q"
    - "ctrl+c"
```

## Approval settings

`ui.approval_open_mode` controls what happens when a normal SSH connect prints a Warpgate approval URL:

- `ask` - prompt locally with open once, always, or no
- `always` - open the approval URL right away
- `never` - never open it automatically

`ui.approval_open_command` is the command template used to open the URL. It must include `%s`, which `wgt` replaces with the approval URL.

Examples:

- `xdg-open %s`
- `open %s`
- `firefox %s`
