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
        # Everything Go here rides `go_latest` rather than a pinned `go_1_NN`,
        # so the repo follows the newest toolchain nixpkgs ships without an
        # edit every release. Bare `pkgs.go` lags a major behind, so the
        # attribute has to be named explicitly.
        goOverlay = _: prev: {
          # Keep the tooling on the same toolchain as the build. go.mod targets
          # 1.27 and these three are built against nixpkgs' default 1.26 —
          # goimports in particular ships wrapped with a `go` on PATH, and when
          # that `go` is older than the go.mod directive it tries to fetch a
          # toolchain from inside the network-less check sandbox.
          # golangci-lint and gopls already track go_latest upstream.
          gofumpt = prev.gofumpt.override { buildGoModule = prev.buildGoLatestModule; };
          gotestsum = prev.gotestsum.override { buildGoModule = prev.buildGoLatestModule; };
          gotools = prev.gotools.override {
            buildGoModule = prev.buildGoLatestModule;
            go = prev.go_latest;
          };
        };
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ goOverlay ];
        };
        fc = flake-checks.lib;
        common = {
          inherit pkgs;
          root = ./.;
          pname = "gososi";
          version = "0.0.1";
          vendorHash = "sha256-U4n57FbAZZ4afMDRexPBvoTgyauTWkohDZozW4WQp64=";
          # flake-checks feeds this to `buildGoModule.override { go = goPkg; }`
          # — which is exactly what buildGoLatestModule is.
          goPkg = pkgs.go_latest;
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
          buildInputs = with pkgs; [ go_latest golangci-lint gofumpt gotestsum gopls gotools prek ];
        };
      }
    );
}
