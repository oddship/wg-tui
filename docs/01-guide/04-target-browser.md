# Target browser walkthrough

## Find and inspect a target

Targets appear in a table above Details by default. In browse mode, Up/Down or `j` / `k` moves the selection; Page Up/Page Down or `ctrl+b` / `ctrl+f` moves one page. Recent targets appear first and have a `↺` marker.

![Target table with Details below](../../_static/assets/wgt-browse-overview.png)

Press `/` and type to search locally. Name, description, group, kind, external host, and default database name are searchable, even when their columns are hidden. Escape leaves search and keeps the filter. Press Escape again to clear it.

The bottom Details pane wraps metadata and descriptions. Use `alt+up` / `alt+down` to scroll through it. The SSH command stays visible; `c` copies the full command even when the preview is shortened.

Enter connects to the selected target. `t` opens the tunnel form and `s` opens the transfer form. Leave search with Escape before using browse actions. Press `?` in browse mode to show the full shortcut list, and press it again to close Help.

## Choose columns and layout

Press `e`, then Tab to Target Columns. The preview above the editor updates as you make changes.

![Column editor and live table preview](../../_static/assets/wgt-target-columns.png)

Use Up/Down to select a column. Space toggles its visibility. `-` moves it up and `+` moves it down. Name stays enabled but can be moved to another position. Tab leaves the column editor for the next setting.

Details Position switches between `bottom` and `right`. Bottom puts Details below the table; right places it beside the table on terminals at least 72 columns wide. Narrower terminals stack the panes in either mode.

Target Metadata Brightness offers `normal`, which uses terminal text color, and `subtle`, which uses a secondary color. On narrow panes, the table hides trailing columns while retaining Name. Full metadata remains available in Details.

Press `ctrl+s` to save from any setting, or Escape to cancel. Display-only changes save locally, including when Warpgate is unavailable. Changes to server or SSH connection settings still require API validation.

## Refresh and offline behavior

Press `r` to refresh. Refresh preserves the selected target by name. If it disappears from the current results, the selection clears. Choose another target before connecting, opening a tunnel, or transferring files.

Fresh cached targets remain usable offline. Stale targets remain available only when `cache.use_stale_on_error` is enabled and the snapshot is within `cache.max_age`. Expired snapshots are discarded and trigger a refresh, even when `cache.ttl` is longer than `cache.max_age`.

See [Configuration](../../reference/configuration/) for the saved fields and cache settings.
