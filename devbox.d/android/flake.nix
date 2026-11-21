{
  # The Android SDK that `mygo build -platform android` and the emulator
  # need: MyGo's Gradle compiles against android-36 with build tools 35,
  # and links Go with the NDK.
  description = "Android SDK, NDK and emulator for Quranku";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "aarch64-darwin" "x86_64-darwin" "x86_64-linux" "aarch64-linux" ];
      forAll = f: nixpkgs.lib.genAttrs systems (system: f (import nixpkgs {
        inherit system;
        config = { allowUnfree = true; android_sdk.accept_license = true; };
      }));
    in {
      packages = forAll (pkgs:
        let
          abi = if pkgs.stdenv.hostPlatform.isAarch64 then "arm64-v8a" else "x86_64";
          android = pkgs.androidenv.composeAndroidPackages {
            platformVersions = [ "36" ];
            buildToolsVersions = [ "35.0.0" ];
            includeNDK = true;
            includeEmulator = true;
            includeSystemImages = true;
            systemImageTypes = [ "google_apis" ];
            abiVersions = [ abi ];
          };
        in {
          default = android.androidsdk;
          android-sdk = android.androidsdk;
        });
    };
}
