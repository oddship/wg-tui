# Configuration

`wgt` stores config in HUML.

## Default locations

By default, `wgt` uses your platform config directory. Cache and local state go in your platform cache directory.

On Linux, that is usually:

- config: `~/.config/wgt/config.huml`
- cache and local state: `~/.cache/wgt/`

The cache directory holds a `last_profile` file and one subdirectory per profile, for example `~/.cache/wgt/prod/`. Each one stores:

- cached Warpgate target snapshots
- recent targets
- remembered transfer form state

## CLI overrides

`wgt` also supports a few small CLI flags:

```bash
wgt --help
wgt --config /path/to/config.huml
wgt --cache-path /path/to/cache-dir
wgt --profile staging
```

`--cache-path` and `--profile` only affect the current run. They do not rewrite `cache.dir` or `default_profile` in your saved config.

## Profiles

Each profile points at one Warpgate upstream and has its own `server` and `ssh` settings. `cache`, `ui` and `keys` are shared by all profiles.

- at startup `wgt` opens `--profile` if passed, otherwise the last profile you used, otherwise `default_profile`
- the last used profile is stored in `<cache.dir>/last_profile`; a profile picked with `--profile` is not remembered
- press `p` in the target list to open the profile picker, then `enter` to switch or `n` to create a new profile
- press `e` to edit, or rename, the active profile
- profile names may only contain letters, digits, `-` and `_`

Configs from older versions with top-level `server` and `ssh` sections still load. They are treated as a profile named `default` and written in the new shape on the next save.

## Example config

You only need to set `cache.dir` if you want to override the platform default.

```huml
%HUML v0.2.0

default_profile: "prod"

profiles::
  prod::
    server::
      url: "https://warpgate.example.com"
      token: "<warpgate-api-token>"
      insecure_skip_tls_verify: false
    ssh::
      username: "user@example.com"
      host: "warpgate.example.com"
      port: 2222
      binary: "ssh"
      extra_args::
        - "-o"
        - "ServerAliveInterval=30"
  staging::
    server::
      url: "https://warpgate-staging.example.com"
      token: "<warpgate-api-token>"
    ssh::
      username: "user@example.com"

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
  page_up::
    - "pgup"
    - "ctrl+b"
  page_down::
    - "pgdown"
    - "ctrl+f"
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
  profile::
    - "p"
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
