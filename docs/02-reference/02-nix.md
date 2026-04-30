# Nix

`wgt` ships a flake with:

- `packages.<system>.default`
- `apps.<system>.default`
- `devShells.<system>.default`

## Run

```bash
nix run github:oddship/wg-tui
```

## Build

```bash
nix build github:oddship/wg-tui
```

## Use in your own flake

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    wg-tui.url = "github:oddship/wg-tui";
  };

  outputs = { self, nixpkgs, wg-tui, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in {
      packages.${system}.default = wg-tui.packages.${system}.default;
    };
}
```

## Add to NixOS or Home Manager

Inside a flake-based setup, add:

```nix
wg-tui.packages.${pkgs.system}.default
```

For example:

```nix
environment.systemPackages = [
  wg-tui.packages.${pkgs.system}.default
];
```
