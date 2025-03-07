{ buildGoModule
, fetchFromGitHub
, lib
}:

buildGoModule rec {
  pname = "licenseupdater";
  version = "2.3.8";

  # Don't run tests.
  doCheck = false;
  doInstallCheck = false;

  src = fetchFromGitHub {
    owner = "redpanda-data";
    repo = "redpanda-operator";
    rev = "v${version}-24.3.6";
    hash = "sha256-/bHiL1VnAyb8oAIYjAOmg68YDymJTktJBk6DE902Q6s=";
  };

  sourceRoot = "source/licenseupdater";

  vendorHash = "sha256-nP2QBuZauE/4+5WdxUWohwnccH7LHV3dPy5EHbkKUgQ=";

  meta = with lib; {
    description = "A small tool for keeping licenses in-sync";
    homepage = "https://github.com/redpanda-data/redpanda-operator/tree/main/licenseupdater";
    mainProgram = "licenseupdater";
  };
}