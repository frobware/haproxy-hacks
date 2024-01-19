{
  description = "A devShell with Python and SSL-related dependencies.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
  flake-utils.lib.eachDefaultSystem (system:
  let
    pkgs = nixpkgs.legacyPackages.${system};
    pythonEnv = pkgs.python3.withPackages (ps: with ps; [
      pyopenssl
    ]);
  in
  {
    devShell = pkgs.mkShell {
      buildInputs = [ pythonEnv pkgs.git ];
      shellHook = ''
        echo "Python with SSL-related packages is ready."
      '';
    };
  }
  );
}

