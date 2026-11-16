# Quickstart

## Install and launch

Download the binary for your platform from [GitHub Releases](https://github.com/oddship/wg-tui/releases/latest). Follow [Install from release](../install-from-release/) for installation commands, or [Nix](../../reference/nix/) for a flake-based setup.

Then run:

```bash
wgt
```

## Set up your first profile

Generate an API token in the Warpgate web UI. On first launch, enter a profile name (initially `default`), server URL, API token, and SSH username. Leave SSH Host blank to use the server URL's hostname; override the SSH port if your deployment needs it.

`wgt` validates the connection and saves the config. On Linux, config usually lives at `~/.config/wgt/config.huml`, with per-profile data under `~/.cache/wgt/`. See [Configuration](../../reference/configuration/) for locations and CLI overrides.

## Find a target and connect

Use Up/Down or `j`/`k` to select a target. Press `/` to search, then Enter to connect with your system SSH client.

Escape leaves search while keeping the filter. Press Escape again to clear it. Leave search before using letter shortcuts such as `t` for a tunnel or `s` for a transfer.

Recently used targets rise to the top and show a `↺` marker. A refresh keeps the selected target by name; if it disappears, choose another before connecting.

## Switch upstreams

Press `p` in browse mode. Use arrow keys and Enter to switch profiles, `n` to create one, or Escape to cancel. The header shows the active name, such as `default (profile)`.

Each profile remembers its own targets, recent targets, and transfer drafts. Startup uses `--profile` when supplied, then a valid last-used profile, then `default_profile`.

## Adjust settings

Press `e` to edit connection settings, rename the profile, or change the target table and Details layout. Use `ctrl+s` to save. Display settings and profile renames work offline; connection changes require API validation. Navigation pauses while saving.

## Next steps

- [Target browser and settings](../target-browser/)
- [SSH approval, tunnels, and transfers](../connections/)
- [Configuration and backup recovery](../../reference/configuration/)
- [Keyboard reference](../../reference/keyboard/)
