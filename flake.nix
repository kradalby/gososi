{
  description = "gososi";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    flake-checks.url = "github:kradalby/flake-checks";
    flake-checks.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs =
    { self
    , nixpkgs
    , flake-utils
    , flake-checks
    , ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        fc = flake-checks.lib;
        common = {
          inherit pkgs;
          root = ./.;
          pname = "gososi";
          version = "0.0.1";
          vendorHash = "sha256-c8ZatcH4hMg8Weoo90xrS8Q0wrEl8t6hScLxLooS4Y8=";
          goPkg = pkgs.go_1_26;
        };
      in
      {
        packages.default = fc.goBuild common;

        formatter = fc.formatter common;

        checks = {
          build = fc.goBuild common;
          gotest = fc.goTest common;
          golangci-lint = fc.goLint common;
          formatting = fc.goFormat common;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [ go_1_26 golangci-lint gotestsum gopls gotools ];
        };
      }
    );
}
