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

Recently used targets rise to the top of the list and show a compact `↺` marker.

If a normal SSH connect prints a Warpgate approval URL, `wgt` can prompt you to open it locally. You can choose open once, always, or no.
