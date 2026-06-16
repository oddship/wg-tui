# Docs capture pipeline

This directory holds the versioned inputs for regenerating `wg-tui` README and published docs media.

## What is versioned here
- `manifest.md` - scenario and asset contract
- `fixtures/targets.json` - fixed target catalog
- `fixtures/state.json` - fixed recent-target and transfer draft state
- `fixtures/config.huml` - fixed config used for previews and forms
- `vhs-theme.json` - pinned terminal theme used by VHS renders

## Recapture

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
- writes final assets to `docs/_static/assets/`
- keeps `PNG` stills plus `GIF`, `MP4`, and `WebM` variants for the animated overview
- renders screenshots and recordings through VHS so the published assets come from a real xterm.js terminal surface instead of ANSI-to-SVG conversion

## Notes
- The product-facing `wgt` CLI stays unchanged. `cmd/wgt-docs` exists only to launch deterministic mocked UI states for docs capture.
- Scenario names, VHS wait patterns, grouping, screenshot filenames, and the tour sequence all come from the single registry in `internal/ui/doc_capture.go`.
- Search and similar flows are replayed through real model key handling, not hand-drawn mockups.
- If the UI changes, update `manifest.md` and the scenario registry together so regenerated assets keep stable names and coverage.
- This directory is intentionally prefixed with `_` so moat skips it during page generation while keeping the capture inputs versioned next to the docs.
