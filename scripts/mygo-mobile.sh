#!/bin/sh
# Runs the MyGo CLI for phones, with Xcode's toolchain:
#
#   devbox run -- scripts/mygo-mobile.sh dev -platform ios
#   devbox run -- scripts/mygo-mobile.sh build -platform ios/simulator
#   devbox run -- scripts/mygo-mobile.sh dev -platform android
#
# devbox's shell compiles C with Nix's clang and macOS SDK, which can't build
# for iOS; this keeps devbox's Go and drops the rest of its environment.
set -eu

go=$(command -v go)
gopath=$(go env GOPATH)
developer=${XCODE_DEVELOPER_DIR:-/Applications/Xcode.app/Contents/Developer}

exec env -i \
	HOME="$HOME" USER="${USER:-}" TERM="${TERM:-xterm-256color}" LANG="${LANG:-en_US.UTF-8}" \
	PATH="$(dirname "$go"):/usr/bin:/bin:/usr/sbin:/sbin:/opt/homebrew/bin" \
	GOPATH="$gopath" GOTOOLCHAIN=local \
	DEVELOPER_DIR="$developer" \
	${ANDROID_HOME:+ANDROID_HOME="$ANDROID_HOME"} \
	${JAVA_HOME:+JAVA_HOME="$JAVA_HOME"} \
	${GOFLAGS:+GOFLAGS="$GOFLAGS"} \
	go tool mygo "$@"
