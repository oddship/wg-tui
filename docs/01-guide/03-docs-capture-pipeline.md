---
title: Docs capture pipeline
description: How the docs screenshots and overview animation are generated and where the versioned inputs live.
---

# Docs capture pipeline

`wg-tui` ships its README and docs media from a reproducible local capture pipeline.

## Published docs assets

The moat site serves checked-in media from `_static/`:

- ![wg-tui overview](../../_static/assets/wgt-overview.gif)
- ![wg-tui tunnel form](../../_static/assets/wgt-tunnel-form.png)
- ![wg-tui transfer form](../../_static/assets/wgt-transfer-form.png)

## Source inputs

The versioned capture inputs live under `docs/_demo/` so they stay close to the docs without becoming published pages:

- `manifest.md` - scenario and asset contract
- `fixtures/targets.json` - fixed target catalog
- `fixtures/state.json` - fixed recent-target and transfer draft state
- `fixtures/config.huml` - fixed config used for previews and forms
- `vhs-theme.json` - pinned terminal theme used by VHS renders

## Regenerate the media

From repo root:

```bash
./scripts/capture-docs.sh
```

If the local machine does not already have the docs toolchain installed, the script re-enters a pinned Nix shell with:

- `vhs`
- `ttyd`
- `ffmpeg`
- `chromium`
- `jetbrains-mono`
- `fontconfig`

The script:

- builds the docs-only helper binary from `cmd/wgt-docs`
- asks that helper to generate the VHS tapes from the authoritative scenario registry in `internal/ui/doc_capture.go`
- writes intermediates under `${BOSUN_WORKSPACE:-workspace}/scratch/wg-tui-doc-capture/`
- writes final published assets to `docs/_static/assets/`
- keeps `PNG` stills plus `GIF`, `MP4`, and `WebM` variants for the animated overview
- renders screenshots and recordings through VHS so the published assets come from a real xterm.js terminal surface instead of ANSI-to-SVG conversion

## Notes

- The product-facing `wgt` CLI stays unchanged. `cmd/wgt-docs` exists only to launch deterministic mocked UI states for docs capture.
- Scenario names, VHS wait patterns, grouping, screenshot filenames, and the tour sequence all come from the single registry in `internal/ui/doc_capture.go`.
- Search and similar flows are replayed through real model key handling, not hand-drawn mockups.
- If the UI changes, update `docs/_demo/manifest.md` and the scenario registry together so regenerated assets keep stable names and coverage.
