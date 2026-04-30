# wgt

`wgt` is a small terminal UI for browsing Warpgate targets, searching them locally, and launching native SSH connections.

It is built for the common Warpgate workflow:
- fetch targets via the Warpgate user API using an API token
- cache targets locally for fast startup and instant search
- pick a target from a searchable list
- launch the system `ssh` client with the right Warpgate username and target format

`wgt` does not implement SSH itself. Warpgate still handles authentication policy, browser approval flows, and connection brokering.

## Features

- TUI-first target picker
- local fuzzy search over cached targets
- cache-first startup with background refresh
- first-run onboarding inside the TUI
- in-app config editing
- configurable keybindings via config
- native `ssh` handoff
- copy SSH command to clipboard

## How it works

`wgt` talks to the Warpgate user API using a token:
- `GET /@warpgate/api/info`
- `GET /@warpgate/api/targets`

When you connect, it launches the system `ssh` binary using Warpgate's target selection format:

```bash
ssh -p <port> -l '<warpgate-username>:<target>' <warpgate-host>
```

If your Warpgate SSH policy requires browser approval or SSO, Warpgate handles that in the SSH session.

## Installation

### Install from GitHub releases

Pick the asset that matches your OS and CPU from the [releases page](https://github.com/oddship/wg-tui/releases).

Current asset names:

- `wgt-linux-amd64`
- `wgt-linux-arm64`
- `wgt-darwin-amd64`
- `wgt-darwin-arm64`

Example for Linux x86_64:

```bash
curl -L -o wgt https://github.com/oddship/wg-tui/releases/latest/download/wgt-linux-amd64
chmod +x wgt
mkdir -p ~/.local/bin
mv wgt ~/.local/bin/
```

Example for Apple Silicon macOS:

```bash
curl -L -o wgt https://github.com/oddship/wg-tui/releases/latest/download/wgt-darwin-arm64
chmod +x wgt
mkdir -p ~/.local/bin
mv wgt ~/.local/bin/
```

Make sure `~/.local/bin` is on your `PATH`, then run:

```bash
wgt
```

### Install with Nix

Build and run from this repo:

```bash
nix run github:oddship/wg-tui
```

Build the package only:

```bash
nix build github:oddship/wg-tui
```

Add it to a flake-based system or Home Manager config:

```nix
inputs.wg-tui.url = "github:oddship/wg-tui";

# packages
inputs.wg-tui.packages.${pkgs.system}.default
```

### Build locally

```bash
go build -o wgt ./cmd/wgt
```

### Run directly

```bash
go run ./cmd/wgt
```

## First run

On first launch, `wgt` opens an onboarding flow and asks for:
- Warpgate server URL
- Warpgate API token
- SSH username
- optional SSH host override
- optional SSH port override

The config is stored in:

```text
~/.config/wgt/config.huml
```

The cache is stored in:

```text
~/.cache/wgt/
```

## Configuration

`wgt` stores config in HUML.

Example:

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
  dir: "~/.cache/wgt"
  ttl: "10m"
  max_age: "168h"
  use_stale_on_error: true

ui:
  details_pane: "right"
  preview_lines: 8

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

## Default keybindings

Browse mode:
- `/` - focus search
- `j` / `k` or arrow keys - move selection
- `enter` - connect to selected target
- `r` - refresh targets from Warpgate
- `e` - edit config
- `c` - copy SSH command
- `?` - toggle help
- `q` - quit

Search mode:
- type to filter locally
- `up` / `down` - move selection while keeping search focused
- `esc` - clear search, then blur search when already empty
- `enter` - connect to selected target
- browse shortcuts do not fire while search is focused

## Notes

- `wgt` expects you to generate a Warpgate API token in the web UI.
- Clipboard support uses `github.com/atotto/clipboard`.
- On Linux, clipboard support may require `xclip` or `xsel` depending on your environment.
- `wgt` is cache-first. If cache exists, it starts quickly and refreshes in the background when stale.

## Development

```bash
go test ./...
go build ./...
```

## Status

This is an early but usable implementation. The main focus is fast target discovery and SSH launch, not full Warpgate API coverage.

## Spec

See [`SPEC.md`](./SPEC.md) for the implementation-oriented product spec.

## License

MIT. See [`LICENSE`](./LICENSE).
