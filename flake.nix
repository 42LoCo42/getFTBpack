{
  outputs = { flake-utils, nixpkgs, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; }; in rec {
        packages.default = pkgs.buildGoModule {
          name = "getFTBpack";
          src = ./.;
          vendorHash = "sha256-cGTT/sUu6f5/1jQqWLh3L5dwD0dZa9N6CZGJSq6Yy/Y=";
        };

        devShells.default = pkgs.mkShell {
          inputsFrom = [ packages.default ];
          packages = with pkgs; [ gopls ];
        };
      });
}
