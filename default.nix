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
    tag = "rel-${finalAttrs.version}";
    hash = "sha256-j962CeP18sHfo0RE5XfT0LXVhibp0BGsifvyYhbxMog=";
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
