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

## First run

On first launch, `wgt` asks for:

- Warpgate server URL
- Warpgate API token
- SSH username
- optional SSH host override
- optional SSH port override

Config is stored at `~/.config/wgt/config.huml`.
Cache is stored under `~/.cache/wgt/`.

## Connect flow

`wgt` uses the Warpgate user API:

- `GET /@warpgate/api/info`
- `GET /@warpgate/api/targets`

When you connect, it launches:

```bash
ssh -p <port> -l '<warpgate-username>:<target>' <warpgate-host>
```
