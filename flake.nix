{
  description = "gososi";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    flake-checks.url = "github:kradalby/flake-checks";
    flake-checks.inputs.nixpkgs.follows = "nixpkgs";
    flake-checks.inputs.flake-utils.follows = "flake-utils";
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
          vendorHash = "sha256-JoRJnSaFKo0DXwPq76cL/o0afIHSjauJK7BxFsLl8j4=";
          goPkg = pkgs.go_1_26;
        };
      in
      {
        packages.default = fc.goBuild common;

        formatter = fc.formatter common;

        checks = {
          build = fc.goBuild common;
          # Nested testdata dirs aren't picked up by the pinned flake-checks
          # source filter (only root ./testdata is); list them so fixtures
          # reach the test sandbox instead of the tests silently skipping.
          gotest = fc.goTest (common // {
            extraSrc = [ ./sosi/testdata ./proj/testdata ];
          });
          golangci-lint = fc.goLint common;
          formatting = fc.goFormat common;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [ go_1_26 golangci-lint gofumpt gotestsum gopls gotools prek ];
        };
      }
    );
}
