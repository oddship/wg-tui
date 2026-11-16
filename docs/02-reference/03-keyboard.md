# Keyboard reference

These are the default bindings. Browse actions can be changed in the HUML `keys` section. The config editor does not edit keybindings; edit the file and restart `wgt`. Conflicting global bindings are rejected. See [Configuration](../configuration/) for the complete example.

## Browse

| Key | Action |
| --- | --- |
| Up/Down or `k`/`j` | Move selection |
| Page Up/Page Down or `ctrl+b`/`ctrl+f` | Move a page |
| `/` | Focus search |
| Escape | Clear the retained search filter |
| Enter | Connect to the selected target |
| `t` | Open the tunnel form |
| `s` | Open the transfer form |
| `p` | Open the profile picker |
| `e` | Open settings |
| `r` | Refresh targets |
| `c` | Copy the selected SSH command |
| `?` | Expand or collapse keyboard help |
| `alt+up` / `alt+down` | Scroll Details in a stacked layout |
| `q` | Quit |

When a refresh removes the selected target, selection clears. Choose another target before connecting, tunneling, or transferring.

## Search

Printable keys edit the query, including browse shortcuts such as `p`, `t`, and `s`. Up/Down moves selection without leaving search; native Page Up/Page Down pages through results. Enter connects. Printable navigation aliases remain text input; `ctrl+f` stays with the search editor for cursor movement. Escape leaves search and retains the query. Escape again in browse mode clears it.

## Profile picker

| Key | Action |
| --- | --- |
| Up/Down | Move selection |
| Enter | Switch to the selected profile |
| `n` | Create a profile |
| Escape | Cancel |

Configured up/down aliases also work when they do not conflict with these local actions. Native arrow keys always remain available.

## Config editor

| Key | Action |
| --- | --- |
| Tab / Shift+Tab or Down/Up | Move between settings |
| Left/Right or `h`/`l` on an option | Change the option |
| `ctrl+s` | Save from any field |
| Enter on an option | Move to the next field |
| Enter on a text field | Save |
| Escape | Cancel before saving |

In Target Columns, Up/Down chooses a column, Space toggles visibility, and `-`/`+` moves it up or down. Name stays enabled. Enter or Tab leaves the column editor.

While saving, further edits and navigation are paused. `ctrl+c` queues an exit after the save finishes.

## Tunnel form

Enter the remote and local ports. Tab/Shift+Tab or Down/Up moves between fields. `c` copies the preview, Enter starts the tunnel, and Escape cancels.

## Transfer form

Tab/Shift+Tab or Down/Up moves between fields. Left/Right or `h`/`l` changes the tool or direction. Enter on an option advances; Enter on a text field runs the transfer. Escape cancels.

## Tunnel page

| Key | Action |
| --- | --- |
| `x` | Close the tunnel |
| `r` | Reconnect |
| `c` | Copy the tunnel command |
| `b` | Close the tunnel and return to browse |
| `?` | Toggle help |
| `q` | Close the tunnel and quit |

`ctrl+c` exits from any TUI page, waiting for an active config save to finish. During an SSH session or transfer, input belongs to the subprocess until it returns.
