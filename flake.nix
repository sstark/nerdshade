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
            version = "1.4.0";
            src = ./.;
            vendorHash = "sha256-7Tdn5YBrsOORsHjlz1r9K/U3AtKdvu80+P3iU3+FXO8=";
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
