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

Each profile points at one Warpgate upstream and has its own `server` and `ssh` settings. If `ssh.host` is omitted, it defaults to the hostname in `server.url`. `cache`, `ui` and `keys` are shared by all profiles.

- at startup `wgt` opens `--profile` if passed, otherwise the last profile you used, otherwise `default_profile`
- the last used profile is stored in `<cache.dir>/last_profile`; a profile picked with `--profile` is not remembered
- press `p` in the target list to open the profile picker, then `enter` to switch or `n` to create a new profile
- press `e` to edit, or rename, the active profile
- profile names may only contain letters, digits, `-` and `_`; `last_profile` is reserved (regardless of case)

In the picker, the arrow keys, `enter`, `n` and `esc` keep their displayed actions even if you use them for custom navigation elsewhere. Other configured up/down aliases also work.

Renaming a profile without changing its connection settings works offline and preserves its cached targets. While a config save is running, navigation and further edits are paused. If you press `ctrl+c`, `wgt` finishes the save before exiting.

Configs from older versions with top-level `server` and `ssh` sections still load. They are treated as a profile named `default` and written in the new shape on the next save.

## Backups and recovery

On the first startup with an older config, `wgt` copies the existing target snapshot and local state into the `default` profile. When you rename a profile, it copies that profile's data to the new name, including recent targets and remembered transfer settings. Target actions are unavailable while a profile switch loads its data.

Before either copy, `wgt` saves the original config and any existing `snapshot.json` and `state.json` files in a dated directory under `<cache.dir>/.backups/`. With `--cache-path`, backups go under that directory instead. The original data stays in place. Backup directories are private, and backup files use owner-only read/write permissions.

A notice stays visible in the target list for the rest of the session, even after a refresh. It shows the backup location and the path to `RESTORE.txt`, which contains the exact source and destination paths for recovery.

To restore, close `wgt`, copy the files listed in `RESTORE.txt` back to their original paths, and move the newly created profile directory aside. Then restart `wgt`. Restoring replaces any config or saved state changed since the backup.

Existing profile directories are never replaced during a copy. Renaming to a name whose data directory already exists stops with an error. Move that directory aside before retrying if you want to reuse the name. If a backup fails, `wgt` retains the original files and reports the error. A failed startup migration disables actions until you exit, fix the error and restart.

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
