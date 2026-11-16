# Contributing

Run commands from the repository root. Use the Go version declared in `go.mod`, or enter the development shell:

```bash
nix develop
```

## Build and run

```bash
go build -o wgt ./cmd/wgt
./wgt --help
```

To run without keeping a binary:

```bash
go run ./cmd/wgt
```

Interactive testing needs a terminal, an SSH client, and access to a Warpgate upstream. `rsync`, `scp`, and clipboard helpers are optional dependencies for their respective features.

## Validate changes

```bash
go test ./...
go test -race ./...
go vet ./...
```

Run the checks relevant to a change. Config and UI tests cover documented examples, profile isolation, cache policy, selected-target identity, save lifecycles, and command construction. Capturing documentation scenarios is opt-in; see the [capture guide](../docs-capture-pipeline/).

Before submitting changes to packaging or dependencies, run the same Nix checks as CI:

```bash
nix flake check --no-build
nix build .#default --no-link
```

`just` is optional. The existing recipes cover build, static build, tests, and running with arguments.

## Keep changes consistent

Read [Architecture](../architecture/) before changing package boundaries or asynchronous behavior. Keep native commands argument-based, profile state isolated, and selection stable by target identity. Add regression coverage when changing connection or persistence behavior.

Use conventional commit subjects such as `fix(ui): ...` or `docs: ...`. Update the user-facing guide or reference when behavior changes, and record unreleased changes in the changelog. Historical release notes and the archived spec stay as history.

## Documentation

The site is built from `docs/` using moat. Numbered directory and file prefixes control ordering; published routes omit those prefixes. Directories beginning with `_` hold capture inputs, static assets, or archived material and do not become pages.

Keep complete config examples in the configuration reference. The example and capture fixture are checked with the config loader. Regenerate media when the visible UI changes; do not add a separate set of manually maintained mock screens.

The Docs workflow builds and deploys pushes to `master` that change `docs/`. Changes to application rendering or capture code also need updated checked-in media to reach the published site.

## Releases

Prepare a versioned changelog entry and reset Unreleased. Merge the tested changes into `master`, then create and push an annotated version tag. The Release workflow builds Linux and macOS binaries for amd64 and arm64, checks the Nix package, and attaches the binaries to a GitHub release. Add the versioned release notes to that release.

Tag publication triggers a release. CI, Docs, and Release are separate workflows; verify the relevant runs before announcing completion.
