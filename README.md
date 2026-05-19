# wgt

`wgt` is a small terminal UI for browsing [Warpgate](https://github.com/warp-tech/warpgate) targets, searching them locally, launching native SSH connections, opening local port-forward tunnels through Warpgate, and running quick file transfers through the selected target.

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
- local tunnel mode with a dedicated status page
- quick `rsync` or `scp` transfers in both directions through Warpgate
- copy SSH or tunnel command to clipboard

## How it works

`wgt` talks to the Warpgate user API using a token:
- `GET /@warpgate/api/info`
- `GET /@warpgate/api/targets`

When you connect, it launches the system `ssh` binary using Warpgate's target selection format:

```bash
ssh -p <port> -l '<warpgate-username>:<target>' <warpgate-host>
```

Tunnel mode starts from the target list inside the TUI, then opens a small port form and starts a native SSH local forward:

```bash
ssh -N \
  -o ExitOnForwardFailure=yes \
  -L <local-port>:127.0.0.1:<remote-port> \
  -p <port> \
  -l '<warpgate-username>:<target>' \
  <warpgate-host>
```

Transfer mode opens a small form with a tool, direction, and local and remote paths.

For `rsync`, `wgt` runs native `rsync` over Warpgate SSH:

```bash
rsync -avz \
  -e "ssh -p <port> -l '<warpgate-username>:<target>' [extra ssh args...]" \
  <local-path> \
  <warpgate-host>:<remote-path>
```

For `scp`, `wgt` runs native `scp` over the same Warpgate target format, with built-in compression and attribute preservation defaults.

Default transfer flags:
- `rsync`: `-avz`
- `scp`: `-C -p`

You can edit these flags directly in the transfer form before running the command.

`wgt` also remembers the last selected transfer tool, direction, flags, local path, and remote path across restarts.

For downloads, the remote and local paths are swapped.

If your Warpgate SSH policy requires browser approval or SSO, Warpgate handles that in the SSH session. During tunnel startup or transfer execution, `wgt` may briefly hand control to `ssh` so you can complete approval before returning to the TUI.

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

### Tunnel mode

Open `wgt`, select a target, then press `t` to open the tunnel form.

The form asks for:
- remote port
- local port

When you type a remote port first, the local port mirrors it until you explicitly change the local port. After submission, `wgt` switches to a dedicated tunnel page that shows status and lets you close or reconnect the tunnel.

### Transfer mode

Open `wgt`, select a target, then press `s` to open the transfer form.

The form asks for:
- tool: select `rsync` or `scp`
- direction: select `upload` or `download`
- flags: editable flags for the selected tool
- local path
- remote path

For uploads, the local path is the source and the remote path is the destination.
For downloads, the remote path is the source and the local path is the destination.

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

## Default keybindings

Browse mode:
- `/` - focus search
- `j` / `k` or arrow keys - move selection
- `enter` - connect to selected target
- `t` - open tunnel form for the selected target
- `s` - open transfer form for the selected target
- `r` - refresh targets from Warpgate
- `e` - edit config
- `c` - copy SSH command
- `?` - toggle help
- `q` - quit

Tunnel form:
- type remote and local ports
- `enter` - start tunnel
- `esc` - cancel and return to the target list

Transfer form:
- use `left` / `right` to choose tool and direction
- edit flags, local path, and remote path directly
- `enter` on a selected option moves to the next field
- `enter` on a text field runs the selected transfer
- `esc` - cancel and return to the target list

Tunnel page:
- `x` - close tunnel
- `r` - reconnect tunnel
- `c` - copy tunnel command
- `b` - close tunnel and return to target list
- `?` - toggle help
- `q` - quit

Search mode:
- type to filter locally
- printable keys continue editing the query, including `t` and `s`
- `up` / `down` - move selection while keeping search focused
- `esc` - blur search and keep the current filter
- `esc` again in browse mode - clear the current filter
- `enter` - connect to selected target
- blur search before using browse actions such as `t` or `s`

## Notes

- `wgt` expects you to generate a Warpgate API token in the web UI.
- Tunnel startup may briefly hand terminal control to `ssh` for approval or login, then return to the TUI once the tunnel is backgrounded.
- Transfer execution may also briefly hand terminal control to SSH-backed authentication or approval flows.
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
