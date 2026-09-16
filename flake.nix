{
  description = "DevOps Sentinel - Development Environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs {
          inherit system;
          config = {
            allowUnfree = true;
          };
        };
      });
    in
    {
      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Backend: PHP 8.3 & Extensions
            (php83.buildEnv {
              extensions = { all, enabled }: with all; enabled ++ [
                pdo_pgsql
                pgsql
                zip
                bcmath
                pcntl
              ];
              extraConfig = ''
                memory_limit = 512M
              '';
            })
            php83Packages.composer

            # Database: PostgreSQL 16 client & tools
            postgresql_16

            # Frontend: Go (hỗ trợ biên dịch WebAssembly GOOS=js GOARCH=wasm)
            go
            gnumake

            # Tiện ích bổ trợ
            git
            curl
            jq
          ];

          shellHook = ''
            echo "🚀 [DevOps Sentinel] Môi trường phát triển đã sẵn sàng!"
            echo "   • Go: $(go version 2>/dev/null || echo 'Chưa nhận diện')"
            echo "   • PHP: $(php -r 'echo PHP_VERSION;' 2>/dev/null || echo 'Chưa nhận diện')"
            echo "   • Composer: $(composer --version 2>/dev/null | head -n 1 || echo 'Chưa nhận diện')"
            echo "   • PostgreSQL CLI: $(psql --version 2>/dev/null || echo 'Chưa nhận diện')"
          '';
        };
      });
    };
}
