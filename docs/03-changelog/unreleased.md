# Unreleased

## Target browser

- Targets now appear in a table above Details by default.
- Column visibility and order can be changed with keyboard controls and a live preview.
- Metadata brightness can be changed in settings, and secondary text has stronger contrast.
- The bottom Details pane wraps full metadata and descriptions, supports scrolling, and keeps the SSH command visible.
- Display-only settings can be saved without contacting the API.

## Fixes

- Refresh preserves the selected target by name. A removed target requires a new selection before a connection, tunnel, or transfer.
- Expanded Help displays browse and tunnel shortcuts within the terminal layout.
- Cache loading honors maximum age and stale fallback settings. Rejected snapshots trigger a refresh even when still within the configured TTL.
- README, spec, and reference config examples use valid HUML and the current display defaults.
