{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        pcs = pkgs.buildGoModule {
          pname = "pcs";
          version = "0.1.0";

          src = ./.;

          vendorHash = null;

          subPackages = [ "cmd/server" ];

          CGO_ENABLED = 0;

          # nativeBuildInputs = [ pkgs.pkg-config ];
          # buildInputs = [ pkgs.stdenv.cc ];
        };
      in
      {
        packages.default = pcs;

        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.gopls
            
            pkgs.google-cloud-sdk
          ];

          shellHook = ''
            export CGO_ENABLED=0
          '';
        };

      }
    );
}
