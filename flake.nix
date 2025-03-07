{
  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    devshell = {
      url = "github:numtide/devshell";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ self
    , devshell
    , flake-parts
    , nixpkgs
    }: flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "aarch64-darwin" "x86_64-linux" "aarch64-linux" ];

      imports = [
        devshell.flakeModule
      ];

      perSystem = { self', system, ... }:
        let
          lib = pkgs.lib;
          pkgs = import nixpkgs {
            inherit system;
            overlays = [
              # Load in various overrides for custom packages and version pinning.
              (import ./support/overlay.nix { pkgs = pkgs; })
            ];
          };
        in
        {
          formatter = pkgs.nixpkgs-fmt;

          devshells.default = {
            env = [
              { name = "PATH"; eval = "$(pwd)/.build:$PATH"; }
            ];

            # If the version of the installed binary is important make sure to
            # update TestToolVersions.
            packages = [
              pkgs.backport
              pkgs.changie # Changelog manager
              pkgs.cobra-cli
              pkgs.gawk # GNU awk, used by some build scripts.
              pkgs.gh
              pkgs.gnused # Stream Editor, used by some build scripts.
              pkgs.go-task
              pkgs.licenseupdater
              pkgs.yq-go
            ];
          };
        };
    };
}