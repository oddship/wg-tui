# Architecture

`wgt` is a Go terminal application built with Bubble Tea, Bubbles, and Lip Gloss. It reads Warpgate's user API and delegates connections to system tools. The archived v1 proposal is design history; this page describes the current code and the rules changes should preserve.

## Package boundaries

| Package | Responsibility |
| --- | --- |
| `cmd/wgt` | Product entry point |
| `internal/app` | CLI options and Bubble Tea program startup |
| `internal/config` | HUML loading, defaults, validation, profiles, and runtime keymaps |
| `internal/api` | HTTP requests and the target/info types used by the UI |
| `internal/cache` | Target snapshots, local state, last-used profile, and backup/copy operations |
| `internal/search` | Local indexing, ranked matching, and recent-target ordering |
| `internal/ssh` | Argument construction and execution for SSH, tunnels, transfers, and approval links |
| `internal/ui` | UI state, rendering, forms, navigation, and asynchronous operation lifecycles |
| `cmd/wgt-docs` | Deterministic UI scenarios for documentation captures |

## Authentication and native tools

The API client requests `/@warpgate/api/info` and `/@warpgate/api/targets`, sending `X-Warpgate-Token`. It has request timeouts and an explicit TLS-verification override. It decodes the fields represented by the Go API types; snapshots do not preserve unknown API fields.

Shell connections and tunnels use the configured SSH binary. Transfers invoke system `rsync` or `scp`. Command builders return an executable and argument slice; the application does not execute them through a shell. Warpgate owns connection authorization, SSO, and approval policy.

Normal SSH connections use a PTY to detect approval URLs and offer local browser opening. The browser command is parsed into executable and arguments. Tunnel processes use a control socket so the TUI can check, close, and reconnect them.

## Profiles and persistence

`Config.Profiles` contains persisted server and SSH settings. `Profile`, `Server`, and `SSH` also hold the active runtime selection. `Use` selects a stored profile; form edits use `CommitActive` to put active settings back into the profile map before saving. Updates clone the map when adding or renaming profiles.

Cache policy, UI settings, and keybindings are shared. Each profile has a separate snapshot and state directory under the effective cache root. `--cache-path` changes that root for one run. The root's `last_profile` file records explicit profile switches; the `--profile` startup override alone is not remembered.

Legacy configs load as `default`. Migration and renaming create a dated backup before copying data, keep originals, and publish the new directory from a temporary staging directory. They refuse to replace existing profile data. Backup failures stop the operation; startup migration failures leave the UI blocked with recovery information. Restore instructions include the original file paths.

## UI ownership and asynchronous work

Model updates own UI state. Commands perform network and filesystem work, then return messages. Long operations should follow an existing lifecycle rather than apply a result to whichever screen happens to be active when it arrives.

- Refresh messages carry both profile name and API-client identity. Connection changes replace that identity; local preference changes can reuse it.
- Profile switches clear old targets and state and block input until the destination cache has loaded.
- Config saves allow one operation at a time, including disk writes. Completion messages carry a generation; stale results are ignored. Navigation and duplicate submissions are blocked while saving. A quit request waits for completion.
- Tunnel startup messages carry a generation and cannot update a replaced session.

Connection edits validate both API endpoints before saving. Display-only edits and name-only renames save offline. Both paths share backup and persistence logic. A local-only save preserves the current target snapshot.

## Cache, search, and selection

Cache visibility is controlled by both TTL and maximum age. Expired snapshots are rejected. Stale snapshots remain usable only when stale fallback is enabled. Rejected cache triggers a refresh even if TTL has not elapsed.

Search runs locally over name, description, group, kind, external host, and default database name. Exact and partial matches outrank fuzzy matches, with recent-target ordering as a secondary preference. Search covers metadata even when its table columns are hidden.

Refreshes preserve selection by target name. If a target disappears, selection clears; an action cannot silently use a replacement row. Profile changes do not carry targets or transfer drafts across upstreams.

## Documentation ownership

The README is the entry point. Guides explain workflows; reference pages own config and shortcuts. Release notes describe shipped changes. The [capture pipeline](../docs-capture-pipeline/) owns screenshot scenarios and overview timing.

The old specification is preserved in `docs/_archive/` and excluded from the published site. It is not an active requirement list. Keep new architectural decisions here and verify behavior with the relevant tests.
