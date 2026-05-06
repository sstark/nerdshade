{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:

let pkgs = import <nixpkgs> { config = {}; overlays = []; }; in

buildGoModule (finalAttrs: {
  pname = "nerdshade";
  version = "1.3.0";

  src = fetchFromGitHub {
    owner = "sstark";
    repo = "nerdshade";
    # tag = "rel-${finalAttrs.version}";
    rev = "55b173da471ab0d8d33b17a39b91e565b0128f2b";
    hash = "sha256-YQ3K2nbuLi5e1ahWIGRLMwgc0R2FC4uLawd6VSwle1M=";
  };

  buildInputs = [ pkgs.bash ];

  vendorHash = "sha256-7Tdn5YBrsOORsHjlz1r9K/U3AtKdvu80+P3iU3+FXO8=";

  ldflags = [ "-s" ];

  meta = {
    description = "Hyprsunset client to smoothly transition color temperature during sunrise/sunset";
    homepage = "https://github.com/sstark/nerdshade";
    license = lib.licenses.mit;
    maintainers = with lib.maintainers; [ ];
    mainProgram = "nerdshade";
  };
})
