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

cache::
  dir: "/path/to/cache/wgt"
  ttl: "10m"
  max_age: "168h"
  use_stale_on_error: true

ui::
  details_pane: "bottom"
  preview_lines: 8
  target_fields: "description, group, kind"
  target_contrast: "normal"
  approval_open_mode: "ask"
  approval_open_command: "xdg-open %s"

keys::
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
  copy::
    - "c"
  help::
    - "?"
  quit::
    - "q"
    - "ctrl+c"
```

## Cache policy

`cache.ttl` determines when targets become stale. Snapshots older than `cache.max_age` are discarded. With `cache.use_stale_on_error: true`, usable stale targets remain available while refreshing and after an API failure. Set it to `false` to hide stale targets until a refresh succeeds. Fresh cached targets remain available in either mode.

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

## Target list display

Press `e` to edit settings. Details Position defaults to `bottom`, with Targets above Details. Choose `right` for side-by-side panes. Narrow terminals stack the panes in either mode.

Targets appear in a table. In Target Columns, use Up/Down to select a column, Space to show or hide it, and `-` / `+` to move it up or down. Name always stays enabled. Tab moves to the next setting.

In the config file, `ui.target_fields` controls the column order: `name`, `description`, `group`, and `kind`. The default is `description, group, kind`, with Name added first. Include `name` explicitly to move it elsewhere, for example `kind, name, group, description`. Use `none` for Name only. Name is always included.

Missing values appear as `-`. Long cells end with `…`. Column widths stay consistent while paging through results. When the pane is too narrow, trailing columns are hidden, with Name always retained. The selected target's details appear in the Details pane. In the bottom pane, use `alt+up` / `alt+down` to scroll through wrapped metadata and descriptions. The SSH command stays visible while scrolling. Recently used names have a `↺` marker.

Target Metadata Brightness defaults to `normal`, which uses the terminal's text color. Choose `subtle` for secondary text that adapts to light and dark backgrounds. Selected rows use the selection color throughout.

A sample table previews changes while you edit the column order or brightness. Press `Ctrl+S` to validate and save. Display-only changes save locally without contacting Warpgate. Changes to connection settings still require API validation. Invalid column names and duplicates show an error before saving. Config files without display settings use these defaults. An existing `details_pane: "right"` setting keeps the side-by-side layout.
