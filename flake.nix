{
  description = "Tiny static web server with TLS and HTTP/2 support";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      forAllSystems =
        f:
        nixpkgs.lib.genAttrs systems (
          system:
          f {
            inherit system;
            pkgs = nixpkgs.legacyPackages.${system};
          }
        );
    in
    {
      packages = forAllSystems (
        { pkgs, ... }:
        rec {
          h2static = pkgs.callPackage ./nix/package.nix { };
          default = h2static;
        }
      );

      nixosModules = rec {
        h2static =
          { pkgs, lib, ... }:
          {
            imports = [ ./nix/module.nix ];
            services.h2static.package = lib.mkDefault self.packages.${pkgs.stdenv.hostPlatform.system}.h2static;
          };
        default = h2static;
      };

      apps = forAllSystems (
        { system, pkgs }:
        rec {
          h2static = {
            type = "app";
            program = nixpkgs.lib.getExe self.packages.${system}.h2static;
            meta = self.packages.${system}.h2static.meta;
          };
          default = h2static;
        }
      );

      devShells = forAllSystems (
        { pkgs, ... }:
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
              gotools
            ];
          };
        }
      );

      formatter = forAllSystems ({ pkgs, ... }: pkgs.nixfmt-tree);
    };
}
