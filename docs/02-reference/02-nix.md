# Nix

The flake supports `x86_64-linux`, `aarch64-linux`, `x86_64-darwin`, and `aarch64-darwin`. Enable Nix's `nix-command` and `flakes` experimental features if your installation requires it.

## Run or build

```bash
nix run github:oddship/wg-tui
nix build github:oddship/wg-tui
```

`nix run` starts the TUI. `nix build` creates `result/bin/wgt`.

## Flake outputs

- `packages.<system>.default`: the application package
- `apps.<system>.default`: the runnable application
- `devShells.<system>.default`: Go development tools
- `devShells.<system>.docs`: screenshot and animation tools

For local development, use `nix develop`. See [Contributing](../../development/contributing/) and the [capture guide](../../development/docs-capture-pipeline/).

## Use in your own flake

This minimal flake re-exports the package for one system:

```nix
{
  inputs.wg-tui.url = "github:oddship/wg-tui";

  outputs = { wg-tui, ... }: {
    packages.x86_64-linux.default = wg-tui.packages.x86_64-linux.default;
  };
}
```

To install it through an existing NixOS or Home Manager setup, add the input to your flake and pass `wg-tui` into the module with `specialArgs` (NixOS) or `extraSpecialArgs` (Home Manager).

A NixOS module can then use:

```nix
{ pkgs, wg-tui, ... }: {
  environment.systemPackages = [
    wg-tui.packages.${pkgs.stdenv.hostPlatform.system}.default
  ];
}
```

In a Home Manager module, use the same package in `home.packages`:

```nix
{ pkgs, wg-tui, ... }: {
  home.packages = [
    wg-tui.packages.${pkgs.stdenv.hostPlatform.system}.default
  ];
}
```
