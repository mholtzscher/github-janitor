{
  description = "github-janitor - A Go CLI tool built with Nix";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };

        version = "0.1.3"; # x-release-please-version

        # Add platform-specific build inputs here (e.g., CGO deps)
        buildInputs = [ ];

        # macOS-specific build inputs for CGO
        darwinBuildInputs = pkgs.lib.optionals pkgs.stdenv.isDarwin [
          pkgs.apple-sdk_15
        ];
      in
      {
        packages.default = pkgs.buildGoApplication {
          pname = "github-janitor";
          inherit version;
          src = ./.;
          modules = ./gomod2nix.toml;
          go = pkgs.go_1_25;

          buildInputs = buildInputs ++ darwinBuildInputs;

          # Set CGO_ENABLED=1 if you need CGO
          CGO_ENABLED = 0;

          ldflags = [
            "-s"
            "-w"
            "-X github.com/mholtzscher/github-janitor/cmd.Version=${version}"
          ];

          meta = with pkgs.lib; {
            description = "A Go CLI tool built with Nix";
            homepage = "https://github.com/mholtzscher/github-janitor";
            license = licenses.mit;
            mainProgram = "github-janitor";
            platforms = platforms.all;
          };
        };

        formatter = pkgs.nixfmt-rfc-style;

        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go_1_25
            pkgs.gopls
            pkgs.golangci-lint
            pkgs.gotools
            pkgs.gomod2nix
            pkgs.just
          ]
          ++ buildInputs
          ++ darwinBuildInputs;

          # Set CGO_ENABLED="1" if you need CGO
          CGO_ENABLED = "0";
        };

        devShells.ci = pkgs.mkShell {
          buildInputs = [
            pkgs.go_1_25
            pkgs.golangci-lint
            pkgs.just
          ]
          ++ buildInputs
          ++ darwinBuildInputs;

          CGO_ENABLED = "0";
        };
      }
    );
}
