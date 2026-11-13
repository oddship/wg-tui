# Quickstart

## Install

### From GitHub releases

Download the matching binary from the releases page and place it somewhere on your `PATH`.

Linux x86_64:

```bash
curl -L -o wgt https://github.com/oddship/wg-tui/releases/latest/download/wgt-linux-amd64
chmod +x wgt
mkdir -p ~/.local/bin
mv wgt ~/.local/bin/
```

Apple Silicon macOS:

```bash
curl -L -o wgt https://github.com/oddship/wg-tui/releases/latest/download/wgt-darwin-arm64
chmod +x wgt
mkdir -p ~/.local/bin
mv wgt ~/.local/bin/
```

Then run:

```bash
wgt
```

### With Nix

Run directly from the flake:

```bash
nix run github:oddship/wg-tui
```

Build the package:

```bash
nix build github:oddship/wg-tui
```

### Build locally

```bash
go build -o wgt ./cmd/wgt
```

Run directly:

```bash
go run ./cmd/wgt
```

Useful flags:

```bash
wgt --help
wgt --config /path/to/config.huml
wgt --cache-path /path/to/cache-dir
```

`--cache-path` only applies to the current run.

## First run

On first launch, `wgt` asks for:

- Warpgate server URL
- Warpgate API token
- SSH username
- optional SSH host override
- optional SSH port override

By default, `wgt` uses your platform config directory. Cache and local state go in your platform cache directory.

On Linux, that is usually:

- config: `~/.config/wgt/config.huml`
- cache and local state: `~/.cache/wgt/`

The cache directory also stores recent targets and remembered transfer form state.

## Browse and connect

`wgt` uses the Warpgate user API:

- `GET /@warpgate/api/info`
- `GET /@warpgate/api/targets`

When you connect, it launches:

```bash
ssh -p <port> -l '<warpgate-username>:<target>' <warpgate-host>
```

Targets appear in a table above Details. Use Up/Down or `j` / `k` to select a target, Page Up/Page Down to move a page, and Enter to connect. Press `/` to search locally. Escape leaves search while keeping the filter; press Escape again to clear it.

The Details pane shows the selected target and SSH command. Use `alt+up` / `alt+down` to scroll long metadata in the bottom pane, and `c` to copy the full SSH command. Press `?` to expand keyboard help.

Press `e` for settings. Target Columns lets you show, hide, and reorder columns with a live preview. Use `ctrl+s` to save. Display-only changes save offline; connection changes still require API validation.

See the [target browser walkthrough](../target-browser/) for the full flow.

Recently used targets rise to the top of the table and show a compact `↺` marker. Refresh keeps the same target selected by name. If that target disappears, use Up/Down to choose another before connecting, tunnelling, or transferring.

If a normal SSH connect prints a Warpgate approval URL, `wgt` can prompt you to open it locally. You can choose open once, always, or no.

If you open the tunnel form with `t`, `wgt` also shows the tunnel command there before launch. Press `c` to copy it from the form or from the tunnel page.
