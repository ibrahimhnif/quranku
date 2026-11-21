#!/bin/sh
# Starts an Android emulator from devbox's SDK, making its virtual device
# the first time:
#
#   devbox run emulator            # a phone (Pixel 8)
#   devbox run emulator tablet     # a tablet (Pixel Tablet)
#
# `devbox run android` then installs and runs the app on it.
set -eu

: "${ANDROID_HOME:?run this with devbox run emulator}"
case "${1:-phone}" in
	tablet) name=Quranku_Tablet device=pixel_tablet ;;
	*) name=Quranku_Phone device=pixel_8 ;;
esac
case "$(uname -m)" in
	arm64 | aarch64) abi=arm64-v8a ;;
	*) abi=x86_64 ;;
esac
image=$(ls -d "$ANDROID_HOME"/system-images/android-*/google_apis/"$abi" | sort -V | tail -1)
package="system-images;$(basename "$(dirname "$(dirname "$image")")");google_apis;$abi"

avdmanager=$(ls "$ANDROID_HOME"/cmdline-tools/*/bin/avdmanager | head -1)
if ! "$ANDROID_HOME/emulator/emulator" -list-avds | grep -qx "$name"; then
	echo "Making the virtual device $name ($package)"
	echo no | "$avdmanager" create avd --name "$name" --package "$package" --device "$device"
	# A keyboard and a GPU of the Mac's, for typing and a smooth screen.
	cfg="${ANDROID_AVD_HOME:-$HOME/.android/avd}/$name.avd/config.ini"
	printf 'hw.keyboard=yes\nhw.gpu.enabled=yes\nhw.gpu.mode=host\n' >> "$cfg"
fi
exec "$ANDROID_HOME/emulator/emulator" -avd "$name" -no-snapshot-save -no-boot-anim
