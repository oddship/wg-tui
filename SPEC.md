# wg-tui Specification

## 1. Document Status

- Name: `wg-tui`
- Primary executable: `wgt`
- Status: Draft v1
- Scope: searchable terminal UI for Warpgate target discovery and SSH launch

## 2. Summary

`wgt` is a terminal user interface for browsing Warpgate targets, searching them locally, inspecting basic target metadata, and launching the native `ssh` client against a selected target.

`wgt` is not a full Warpgate client. Warpgate remains responsible for authentication policy, authorization, auditing, connection brokering, and SSH-side browser approval flows.

## 3. Goals

The product MUST:
- start quickly
- make target discovery fast
- support local fuzzy search over cached target data
- launch native SSH for the selected target
- support first-run onboarding inside the TUI
- store configuration in HUML
- allow configuration review and editing from inside the TUI
- support manual cache refresh from the TUI

The product SHOULD:
- render cached targets immediately when available
- refresh stale cache in the background
- remain usable when the API is temporarily unavailable but cache exists
- highlight matched search characters in the target list

## 4. Non-goals

The product MUST NOT:
- implement SSH in Go
- replace Warpgate's browser-based SSH auth flow
- manage Warpgate admin configuration
- aim for full Warpgate OpenAPI coverage
- require a large command-heavy CLI surface for v1

## 5. Primary User and Workflow

### 5.1 Primary user

The primary user:
- already has Warpgate web access
- can generate a Warpgate API token in the web UI
- mainly needs a fast searchable launcher for targets

### 5.2 Primary workflow

1. User generates a Warpgate API token in the Warpgate web UI.
2. User launches `wgt`.
3. If config does not exist, `wgt` runs onboarding inside the TUI.
4. `wgt` loads cached data if present.
5. `wgt` fetches or refreshes `/info` and `/targets` as needed.
6. User searches targets locally.
7. User selects a target and presses the configured connect key.
8. `wgt` launches native `ssh` using Warpgate's username-plus-target format.
9. If Warpgate requires browser approval or SSO for SSH, Warpgate handles that within the SSH session.

## 6. Product Requirements

### 6.1 Executable and packaging

- The installed executable MUST be named `wgt`.
- The repository MAY remain named `wg-tui`.
- The application MUST be TUI-first.

### 6.2 Core user experience

The main experience MUST support:
- opening the app with `wgt`
- viewing cached targets immediately when available
- searching as the user types
- inspecting selected target metadata in a detail pane
- launching SSH for the selected target
- manually refreshing cached data
- opening a config screen from within the TUI

## 7. Authentication and Transport

### 7.1 API authentication

The application MUST use a Warpgate API token for HTTP API requests.

Supported v1 endpoints:
- `GET /@warpgate/api/info`
- `GET /@warpgate/api/targets`

The application MUST send:
- `X-Warpgate-Token: <token>`

The application MUST NOT depend on browser cookie handoff for API authentication.

### 7.2 SSH transport

The application MUST shell out to the system `ssh` binary.

The application MUST assume the Warpgate SSH username format:
- `<warpgate-username>:<target-name>`

The application SHOULD prefer an argument-safe invocation shape such as:

```bash
ssh -p 2222 -l '<warpgate-user>:<target>' <warpgate-host>
```

The implementation MUST avoid shell string interpolation.

## 8. Configuration Specification

### 8.1 Format and location

Configuration MUST be stored in HUML.

Default config file:
- `~/.config/wgt/config.huml`

Default cache directory:
- `~/.cache/wgt/`

### 8.2 HUML syntax constraints

The implementation MUST follow the HUML syntax accepted by koanf's HUML parser.

Relevant syntax rules:
- scalars use `key: value`
- arrays use either block form:
  - `key::` then `- item`
- or inline form:
  - `key:: a, b, c`
- empty arrays use `key:: []`
- empty maps use `key:: {}`
- strings SHOULD be quoted when they contain special characters, URLs, usernames, or durations

When the app writes config back to disk, it SHOULD prefer readable block-list output where appropriate.

### 8.3 Required config fields

The config model MUST support at least:
- server URL
- API token
- TLS insecure toggle
- SSH username
- SSH host override
- SSH port override
- SSH binary path or command name
- SSH extra args
- cache directory
- cache TTL
- cache max age
- stale-on-error toggle
- basic UI preferences
- keybindings

### 8.4 Example config

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

### 8.5 Config loading and persistence

The implementation SHOULD use:
- `github.com/knadh/koanf/v2` for configuration orchestration
- `github.com/knadh/koanf/parsers/huml` for HUML parsing and marshaling
- a koanf file provider for reading and writing config files

The intended flow is:
1. read config file
2. load it into koanf using the HUML parser
3. unmarshal koanf into the application config struct
4. when saving, marshal config back through koanf's HUML parser and write it to disk

## 9. Onboarding Requirements

If the config file does not exist, the application MUST show onboarding instead of the main target browser.

Onboarding MUST:
1. prompt for server URL
2. prompt for SSH username
3. prompt for API token
4. allow optional SSH host override
5. allow optional SSH port override
6. validate connectivity using `/info` and `/targets`
7. write the initial HUML config file
8. transition into the main UI on success

Onboarding SHOULD:
- show actionable validation errors
- avoid requiring manual file editing on first run

## 10. In-App Config Management

The TUI MUST provide a config view or editor.

The config UI MUST allow the user to:
- open config from the main screen
- inspect effective config values
- edit and save config values
- edit keybindings
- reload config after save

The config editor MAY start as a structured form rather than a raw text editor.

## 11. API Data Model

### 11.1 Target fields

The implementation MUST handle at least these target fields:
- `name`
- `description`
- `kind`
- `external_host`
- `group.name`
- `default_database_name`

The implementation SHOULD preserve the full API payload in cache for compatibility with future UI enhancements.

### 11.2 Info fields

The implementation SHOULD read `/info` to derive:
- Warpgate version if useful for display
- SSH port when available
- other metadata only if it improves UX

## 12. Search Specification

### 12.1 Search model

Search MUST be local and in-memory.

Search MUST NOT call the API on every keystroke.

### 12.2 Search input fields

The search index MUST include at least:
- target name
- description
- group name
- kind
- external host

### 12.3 Fuzzy library

The implementation SHOULD use:
- `github.com/sahilm/fuzzy`

Reasoning:
- small
- well known
- local-only
- exposes ranking and matched indexes suitable for TUI highlighting

### 12.4 Search algorithm

The implementation SHOULD use a name-first ranking strategy.

Recommended ranking order:
1. exact name match
2. name prefix match
3. fuzzy match on name
4. fuzzy match across secondary indexed fields
5. weaker fallback matches

The implementation SHOULD maintain at least:
- a primary search string focused on name
- a secondary search string combining description, group, kind, and external host

## 13. Caching Specification

### 13.1 General requirements

Target lists may be large. Caching is a core requirement.

The application MUST persist cache to disk.
The application MUST search cached data locally.

### 13.2 Cache strategy

The cache strategy MUST be stale-while-revalidate.

Startup behavior MUST be:
1. load cache immediately if present
2. render cached targets immediately
3. determine whether cache is stale using TTL
4. refresh in the background when stale
5. replace in-memory data after successful refresh
6. persist fresh data to disk

If refresh fails, the application MUST:
- continue using cache when allowed by `max_age`
- indicate stale state in the UI

### 13.3 Manual refresh

The TUI MUST support manual refresh regardless of cache age.

Manual refresh MUST:
- be invokable via a configurable keybinding
- force a network fetch attempt
- update cache on success
- report failure without discarding existing usable cache

### 13.4 Cache contents

The cache SHOULD persist:
- fetched timestamp
- server identity
- target payload
- info payload

Suggested files:
- `targets.json`
- `info.json`
- `meta.json`

Cache identity SHOULD include at least:
- server base URL
- SSH username if future personalization requires it

## 14. TUI Specification

### 14.1 Technical stack

Preferred initial stack:
- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/bubbles`
- `github.com/charmbracelet/lipgloss`

### 14.2 Main screen layout

The main screen MUST include:
- search input
- target list
- details pane
- footer or status line

### 14.3 List content

Each row SHOULD show compact metadata, such as:
- target name
- target kind
- target group

### 14.4 Details pane content

The details pane SHOULD show:
- name
- description
- kind
- group
- external host
- default database name when present
- last refresh time
- stale/fresh indicator

## 15. Keybinding Specification

### 15.1 General requirements

Keybindings MUST be configurable through the HUML config file.
The app MUST NOT hardcode final user-facing key choices in update logic.

### 15.2 Implementation model

The implementation SHOULD use:
- `github.com/charmbracelet/bubbles/key`
- `key.Binding`
- `key.Matches(...)`

The implementation SHOULD:
- define semantic actions rather than raw keys
- build a runtime keymap from config
- use the same keymap for update logic and help rendering
- rebuild the keymap after config changes where practical

### 15.3 Required semantic actions

The config model MUST support keybindings for at least:
- move up
- move down
- focus search
- clear or unfocus search
- connect to selected target
- refresh from API
- open config editor
- copy SSH command
- toggle help
- quit

### 15.4 Validation

The implementation SHOULD validate keybindings and reject obvious collisions for incompatible global actions.

The implementation MAY keep a safe emergency quit such as `ctrl+c` even if not user-configurable.

## 16. SSH Launch Behavior

When the user selects a target, the application MUST:
1. resolve SSH host and port
2. construct the username as `<configured-username>:<target-name>`
3. invoke the system `ssh` binary directly

The implementation MUST pass arguments without shell interpolation.

The implementation SHOULD verify the safest OpenSSH invocation for usernames containing `@`.
Using `-l` is currently the preferred direction.

## 17. HTTP Client Requirements

The HTTP client MUST:
- support custom base URL
- send `X-Warpgate-Token`
- support TLS with an explicit insecure toggle
- use request timeouts for startup fetch and refresh
- return useful errors for 401, 403, connectivity failures, and malformed config
- expose a manual refresh path used by the refresh keybinding

## 18. Dependency Guidance

Preferred initial dependencies:
- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/bubbles`
- `github.com/charmbracelet/lipgloss`
- `github.com/knadh/koanf/v2`
- `github.com/knadh/koanf/parsers/huml`
- `github.com/sahilm/fuzzy`
- `github.com/atotto/clipboard` for optional text-only clipboard support

## 19. Minimal v1 API Surface

v1 MUST require only:
- `GET /info`
- `GET /targets`

Future additions MAY include:
- SSO provider discovery
- profile inspection
- token introspection if Warpgate exposes something useful

## 20. Proposed Package Layout

```text
cmd/wgt/
  main.go

internal/app/
  run.go

internal/config/
  config.go
  huml.go
  keymap.go

internal/api/
  client.go
  targets.go
  info.go

internal/cache/
  cache.go

internal/search/
  index.go
  rank.go
  fuzzy.go

internal/ssh/
  launch.go

internal/ui/
  model.go
  keys.go
  styles.go
  views.go
  onboarding.go
  config_editor.go
  help.go
```

## 21. Milestones

### 21.1 Milestone 1

- first-run onboarding
- HUML config load/save
- configurable keybindings
- API client for `/info` and `/targets`
- disk cache
- target list
- local fuzzy search
- manual refresh
- config view/edit screen
- SSH launch

### 21.2 Milestone 2

- stale/fresh indicators
- improved ranking polish
- clipboard support
- favorites or recents
- group and kind filters

### 21.3 Milestone 3

- install instructions
- release binaries
- packaged executable name `wgt`

## 22. Open Questions

1. What is the safest OpenSSH invocation when the Warpgate username itself contains `@`?
2. Should cache-first startup remain the default even on very fast networks?
3. Should favorites and recents remain out of v1?
4. Should remote search fallback ever be added for very large target inventories?
5. How strict should keybinding collision validation be in v1?

## 23. Acceptance Criteria for v1

v1 is acceptable when all of the following are true:
- `wgt` starts and shows onboarding when config is missing
- onboarding can save a valid HUML config
- `wgt` can authenticate to Warpgate using an API token
- `wgt` can load and cache `/targets`
- `wgt` can render cached targets on startup
- `wgt` can search targets locally using fuzzy matching
- `wgt` can refresh targets manually from the UI
- `wgt` can open and save config from inside the TUI
- `wgt` can launch native SSH for the selected target
- keybindings are read from config rather than being hardwired
