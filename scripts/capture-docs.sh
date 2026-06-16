#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKSPACE_ROOT="${BOSUN_WORKSPACE:-$ROOT/workspace}"
if [[ "$WORKSPACE_ROOT" != /* ]]; then
  WORKSPACE_ROOT="$ROOT/$WORKSPACE_ROOT"
fi

SCRATCH_DIR="${WG_TUI_CAPTURE_SCRATCH:-$WORKSPACE_ROOT/scratch/wg-tui-doc-capture}"
BIN_DIR="$SCRATCH_DIR/bin"
TAPE_DIR="$SCRATCH_DIR/tapes"
ASSET_DIR="$ROOT/docs/_static/assets"
ASSET_DIR_REL="docs/_static/assets"
SCRATCH_DIR_REL=""
DOCS_BIN="$BIN_DIR/wgt-docs"

detect_browser() {
  if [[ -n "${CHROME_PATH:-}" && -x "${CHROME_PATH}" ]]; then
    printf '%s\n' "$CHROME_PATH"
    return 0
  fi

  if [[ -n "${BROWSER:-}" ]]; then
    local browser_cmd
    read -r browser_cmd _ <<<"$BROWSER"
    if command -v "$browser_cmd" >/dev/null 2>&1; then
      command -v "$browser_cmd"
      return 0
    fi
  fi

  local candidate
  for candidate in chromium chromium-browser google-chrome google-chrome-stable chrome; do
    if command -v "$candidate" >/dev/null 2>&1; then
      command -v "$candidate"
      return 0
    fi
  done

  return 1
}

has_capture_tools() {
  command -v go >/dev/null 2>&1 && command -v python3 >/dev/null 2>&1 && command -v vhs >/dev/null 2>&1 && command -v ttyd >/dev/null 2>&1 && command -v ffmpeg >/dev/null 2>&1 && detect_browser >/dev/null 2>&1
}

ensure_capture_environment() {
  local browser_path=""
  if browser_path="$(detect_browser 2>/dev/null)"; then
    export PATH="$(dirname "$browser_path"):$PATH"
  fi

  if has_capture_tools; then
    return
  fi

  if command -v nix >/dev/null 2>&1; then
    exec nix --extra-experimental-features 'nix-command flakes' develop "$ROOT#docs" -c env WG_TUI_CAPTURE_IN_NIX=1 bash "$0" "$@"
  fi

  echo "Missing docs capture dependencies. Install go, python3, vhs, ttyd, ffmpeg, and a Chromium-compatible browser, or run with nix available." >&2
  exit 1
}

main() {
  ensure_capture_environment "$@"

  SCRATCH_DIR_REL="$(python3 -c 'import os,sys; print(os.path.relpath(sys.argv[1], sys.argv[2]))' "$SCRATCH_DIR" "$ROOT")"

  mkdir -p "$SCRATCH_DIR" "$BIN_DIR" "$TAPE_DIR" "$ASSET_DIR"
  rm -f "$ASSET_DIR"/wgt-*.svg
  rm -f "$ASSET_DIR"/wgt-*.png "$ASSET_DIR"/wgt-overview.gif "$ASSET_DIR"/wgt-overview.mp4 "$ASSET_DIR"/wgt-overview.webm
  rm -f "$SCRATCH_DIR"/extras.gif

  echo "Building docs capture helper into $DOCS_BIN"
  (
    cd "$ROOT"
    go build -o "$DOCS_BIN" ./cmd/wgt-docs
  )

  local overview_tape="$TAPE_DIR/overview.tape"
  local extras_tape="$TAPE_DIR/extras.tape"
  "$DOCS_BIN" --write-vhs-tapes "$TAPE_DIR" --binary "$DOCS_BIN" --asset-dir "$ASSET_DIR_REL" --scratch-dir "$SCRATCH_DIR_REL"

  echo "Rendering overview media with VHS"
  (
    cd "$ROOT"
    vhs "$overview_tape"
  )

  echo "Rendering supporting screenshots with VHS"
  (
    cd "$ROOT"
    vhs "$extras_tape"
  )

  rm -f "$SCRATCH_DIR"/extras.gif

  echo
  echo "Generated assets:"
  find "$ASSET_DIR" -maxdepth 1 \( -name 'wgt-*.png' -o -name 'wgt-overview.*' \) | sort
}

main "$@"
