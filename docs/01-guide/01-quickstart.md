# Quickstart

## Install

Build locally:

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
