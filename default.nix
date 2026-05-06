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
    rev = "pure-testing";
    hash = "sha256-SQjkh1nQgWVcn30rYWnmh62suHzidZBh4fLJCueJWEU=";
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
