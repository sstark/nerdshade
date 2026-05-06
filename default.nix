{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:

buildGoModule (finalAttrs: {
  pname = "nerdshade";
  version = "1.3.0";

  src = fetchFromGitHub {
    owner = "sstark";
    repo = "nerdshade";
    # tag = "rel-${finalAttrs.version}";
    rev = "603ea46c560da92e08e3d968e6407e20a2366d3f";
    hash = "sha256-hVBIw4ZOs/MXskC/TVP3yju0Ucb1rHeCvZfuq5W+ixQ=";
  };

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
