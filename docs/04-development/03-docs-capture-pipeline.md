---
title: Docs capture pipeline
description: Regenerate terminal screenshots and the overview animation from deterministic UI scenarios.
---

# Docs capture pipeline

The checked-in media in `docs/_static/assets/` comes from real xterm.js terminal captures driven by VHS. `cmd/wgt-docs` supplies deterministic UI states without connecting to a live Warpgate server.

## Inputs and outputs

`internal/ui/doc_capture.go` is the scenario registry. It owns scene names, model builders, wait patterns, output filenames, ordering, and timing. Versioned inputs live in `docs/_demo/`:

- `fixtures/config.huml`: mock profiles and settings
- `fixtures/targets.json`: target catalog
- `fixtures/state.json`: recent targets and transfer drafts
- `vhs-theme.json`: terminal theme
- `manifest.md`: the expected asset contract

The [asset manifest](https://github.com/oddship/wg-tui/blob/master/docs/_demo/manifest.md) lists the screenshots and overview formats. Update it and the registry together when adding or removing scenes.

## Regenerate media

From the repository root:

```bash
./scripts/capture-docs.sh
```

The script uses installed tools when available; otherwise it enters the pinned `nix develop .#docs` shell. The toolchain includes Go, Python, VHS, ttyd, FFmpeg/ffprobe, Chromium, JetBrains Mono, and font configuration.

Intermediates go under `${BOSUN_WORKSPACE:-workspace}/scratch/wg-tui-doc-capture/`. To use another scratch directory:

```bash
WG_TUI_CAPTURE_SCRATCH=/tmp/wgt-doc-media ./scripts/capture-docs.sh
```

The script builds the docs helper, generates VHS tapes and `overview.ffconcat`, captures screenshots, and writes PNG, GIF, MP4, and WebM outputs to `docs/_static/assets/`. The default `workspace/` directory is ignored by Git.

## Pacing and validation

Each overview screen stays visible for five seconds, making a 25-second loop. FFmpeg encodes the screenshots with explicit timestamps from the generated manifest. Missed recorder frames cannot shorten those pauses. The script checks all three exported durations against the timeline and fails if they differ by more than 0.2 seconds.

To validate fixtures and write ANSI scenes without VHS:

```bash
WG_TUI_CAPTURE_OUT=/tmp/wgt-doc-captures go test ./internal/ui -run TestDocCaptureScenarios -count=1
```

This covers the same models as the screenshots, including the profile picker and column editor. Search flows replay model key handling. After regenerating media, inspect the affected screenshots and play the overview; an ANSI capture alone does not verify video timing or image layout.

Commit the relevant source changes and final assets together. Keep helper binaries, tapes, and other intermediates out of version control.
