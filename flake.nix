{
  description = "wgt - Warpgate target picker for the terminal";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          default = pkgs.buildGoModule {
            pname = "wgt";
            version = "0.0.1";
            src = self;
            subPackages = [ "cmd/wgt" ];
            vendorHash = "sha256-Lky4soZeNEnuLZiQpRRr+M9/+UTdty4HS2xo41L+8CA=";

            meta = with pkgs.lib; {
              description = "Warpgate target picker for the terminal";
              homepage = "https://github.com/oddship/wg-tui";
              license = licenses.mit;
              mainProgram = "wgt";
              platforms = platforms.unix;
            };
          };
        });

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/wgt";
        };
      });

      devShells = forAllSystems (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
            ];
          };
        });
    };
}
