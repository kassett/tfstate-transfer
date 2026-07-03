{
  description = "Development environment for tfstate-transfer";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      systems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-darwin"
        "x86_64-linux"
      ];
      forEachSystem = nixpkgs.lib.genAttrs systems;
    in
    {
      devShells = forEachSystem (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            config.allowUnfreePredicate = pkg:
              builtins.elem (nixpkgs.lib.getName pkg) [ "terraform" ];
          };
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              docker-compose
              git
              go
              golangci-lint
              gopls
              gotools
              nodejs_22
              pnpm
              terraform
            ];

            shellHook = ''
              export GOTOOLCHAIN=local
              export GOPATH="$PWD/.go"
              export GOCACHE="$PWD/.cache/go-build"
              export GOMODCACHE="$PWD/.cache/go-mod"
              export PATH="$GOPATH/bin:$PATH"
            '';
          };
        });

      formatter = forEachSystem (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        pkgs.nixpkgs-fmt);
    };
}
