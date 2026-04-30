# Install from release

`wgt` release binaries are attached to each GitHub release.

## Pick the right asset

- Linux x86_64: `wgt-linux-amd64`
- Linux ARM64: `wgt-linux-arm64`
- macOS Intel: `wgt-darwin-amd64`
- macOS Apple Silicon: `wgt-darwin-arm64`

## Linux example

```bash
curl -L -o wgt https://github.com/oddship/wg-tui/releases/latest/download/wgt-linux-amd64
chmod +x wgt
mkdir -p ~/.local/bin
mv wgt ~/.local/bin/
```

## macOS Apple Silicon example

```bash
curl -L -o wgt https://github.com/oddship/wg-tui/releases/latest/download/wgt-darwin-arm64
chmod +x wgt
mkdir -p ~/.local/bin
mv wgt ~/.local/bin/
```

## Add to PATH

If `~/.local/bin` is not already on your `PATH`, add this to your shell config:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Then reload your shell and run:

```bash
wgt
```
