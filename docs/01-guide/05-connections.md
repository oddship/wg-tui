# SSH, tunnels, and transfers

Select a target in browse mode before starting a connection. Leave search with Escape so letter shortcuts apply to browse actions.

## Connect with SSH

Press Enter to hand the terminal to your system SSH client. `wgt` uses Warpgate's username-plus-target format:

```bash
ssh -p <port> -l '<warpgate-username>:<target>' <warpgate-host>
```

The Details pane shows the command; `c` copies it. The configured SSH binary and extra arguments are used for shell connections and tunnels. `rsync` uses that binary as its remote shell; `scp` uses the system `scp` command with the configured port, username, host, and extra SSH arguments.

Warpgate remains responsible for SSH authorization and approval. If a normal shell connection prints an approval URL, `wgt` can offer to open it locally: open once, always, or no. Choosing always saves the shared browser preference. See [Approval settings](../../reference/configuration/#approval-settings) for browser commands, including macOS setup.

Tunnel and transfer startup may hand control to SSH so you can finish approval, SSO, or host-key prompts.

## Open a tunnel

Press `t` to open the tunnel form.

![Tunnel form](../../_static/assets/wgt-tunnel-form.png)

Enter the remote service port and a local port. The local value follows the remote value until you change it yourself. The form previews the SSH command; `c` copies it, including placeholder ports before the form is complete.

Enter starts a native local forward and opens the tunnel status page. The current command form is:

```bash
ssh -N \
  -o ExitOnForwardFailure=yes \
  -L <local-port>:127.0.0.1:<remote-port> \
  -p <port> \
  -l '<warpgate-username>:<target>' \
  <warpgate-host>
```

On the tunnel page, `x` closes the tunnel, `r` reconnects, and `c` copies the command. `b` closes the tunnel and returns to the target list. `q` closes it and quits. One tunnel is managed at a time.

## Transfer files

Press `s` to open the transfer form.

![Transfer form](../../_static/assets/wgt-transfer-form.png)

Choose `rsync` or `scp`, select upload or download, and enter local and remote paths. For upload, the local path is the source. For download, the remote path is the source.

Default flags are `-avz` for `rsync` and `-C -p` for `scp`. Edit them in the form before running. Enter on an option moves to the next field; Enter on a text field runs the transfer.

`wgt` remembers the selected tool, direction, flags, and paths per profile. Flags are split on whitespace; the flags field is not a shell command. Local and remote paths have their own fields.

For example, an `rsync` upload uses:

```bash
rsync -avz \
  -e "ssh -p <port> -l '<warpgate-username>:<target>'" \
  <local-path> \
  <warpgate-host>:<remote-path>
```

Install `rsync` or `scp` on your machine before using the corresponding tool. The target must support the chosen transfer protocol.

## Clipboard support

On Linux, copying commands may require `xclip` or `xsel`. A clipboard failure is reported in the status bar; it does not prevent connecting.

See the [keyboard reference](../../reference/keyboard/) for all mode-specific controls.
