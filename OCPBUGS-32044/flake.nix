{
  description = "A Nix flake for $PWD.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, nixpkgs }: let
    forAllSystems = function: nixpkgs.lib.genAttrs [
      "aarch64-linux"
      "x86_64-linux"
    ] (system: function system);
  in {
    devShells = forAllSystems (system: let
      pkgs = import nixpkgs {
        inherit system;
      };
    in {
      default = pkgs.mkShell {
        buildInputs = [
          pkgs.openjdk11
          pkgs.maven
          pkgs.eclipses.eclipse-java
          pkgs.bashInteractive
        ];

        shellHook = ''
        '';
      };
    });
  };
}
