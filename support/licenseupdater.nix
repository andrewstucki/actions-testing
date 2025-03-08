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
    rev = "adb7a5bf652791b2451edaff0f9b87bc14796ef4";
    hash = "sha256-ndu9Aql1TsweAFEdqrEf/mHLbO1moaLvqEvVP/XhCuM=";
  };

  sourceRoot = "source/licenseupdater";

  vendorHash = "sha256-nP2QBuZauE/4+5WdxUWohwnccH7LHV3dPy5EHbkKUgQ=";

  meta = with lib; {
    description = "A small tool for keeping licenses in-sync";
    homepage = "https://github.com/redpanda-data/redpanda-operator/tree/main/licenseupdater";
    mainProgram = "licenseupdater";
  };
}