{
  description = "Hyprsunset client to smoothly transition color temperature during sunrise/sunset";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.buildGoModule {
            pname = "nerdshade";
            version = "1.4.1";
            src = ./.;
            vendorHash = "sha256-xDrdGCKnSP7EgrxqkiMFiwPG20dkOf3rer8lXegKKGo=";
            ldflags = [ "-s" ];
            meta = {
              description = "Hyprsunset client to smoothly transition color temperature during sunrise/sunset";
              homepage = "https://github.com/sstark/nerdshade";
              license = pkgs.lib.licenses.mit;
              mainProgram = "nerdshade";
            };
          };
        });

      checks = forAllSystems (system: {
        build = self.packages.${system}.default;
      });

      devShells = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system}; in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [ go gopls gotools ];
          };
        });
    };
}
