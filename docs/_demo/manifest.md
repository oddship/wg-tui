# wg-tui docs capture manifest

This directory defines the versioned inputs and expected outputs for the local docs capture pipeline.

## Shared assumptions
- Scenario width: `120` columns
- Scenario height: `34` rows
- Authoritative scenario registry: `internal/ui/doc_capture.go`
- Render surface: VHS driving `ttyd` and capturing the xterm.js canvas in Chromium
- Overview layout: `JetBrains Mono`, `19px`, line height `1.05`, `1600x860`
- Form layout: `JetBrains Mono`, `17px`, line height `1.05`, `1600x1100`
- Theme: `vhs-theme.json`
- Intermediates: `${BOSUN_WORKSPACE:-workspace}/scratch/wg-tui-doc-capture/`
- Final checked-in assets: `docs/_static/assets/`

## Fixture inputs
- `fixtures/config.huml` - deterministic config for command previews and forms
- `fixtures/targets.json` - fixed target catalog used across browse, search, tunnel, and transfer states
- `fixtures/state.json` - fixed recent-target and transfer draft state

## Static assets
- `wgt-browse-overview.png`
- `wgt-search-prod.png`
- `wgt-onboarding.png`
- `wgt-config-editor.png`
- `wgt-tunnel-form.png`
- `wgt-transfer-form.png`
- `wgt-tunnel-status.png`

## Animation outputs
- `wgt-overview.gif` - current README/docs embed
- `wgt-overview.mp4` - maintained for higher-quality sharing and future docs embeds
- `wgt-overview.webm` - maintained for browser-friendly video use outside the current markdown embeds

## Scenario contract
- Single-screen scenarios are defined once in `internal/ui/doc_capture.go` and exposed by `cmd/wgt-docs --list-scenarios`.
- Current stills cover browse, search, onboarding, config editor, tunnel form, transfer form, and tunnel status.
- `tour-sequence` is derived from the registry's ordered overview scenes: browse overview -> search results -> tunnel form -> transfer form -> tunnel status.

## Tape behavior
- `cmd/wgt-docs --write-vhs-tapes` generates the VHS tapes directly from the scenario registry, so the script does not maintain a second hand-written list.
- The overview render hides shell setup between scenes so the published animation only shows the TUI states.
- The same overview tape also captures the browse, search, tunnel, transfer, and tunnel-status PNG stills.
- A second pass captures onboarding and config-editor PNG stills.

If the UI changes, update this manifest first, then regenerate assets with `./scripts/capture-docs.sh`.

This directory is prefixed with `_` so moat skips it during page generation.
