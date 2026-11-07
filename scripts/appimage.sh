#!/bin/sh
# Packs a Linux build of `mygo build` as an AppImage.
#
#   scripts/appimage.sh build/linux-amd64 x86_64
#
# The AppImage holds the app, its desktop entry and icon. It uses the
# system's GTK 3, as the app's Debian package does. Needs appimagetool on
# the PATH (https://github.com/AppImage/appimagetool).
set -eu

dir=$1  # build/linux-amd64
arch=$2 # x86_64 or aarch64
version=$(sed -n 's/.*"version": *"\([^"]*\)".*/\1/p' mygo.json)

appdir=$(mktemp -d)/Quranku.AppDir
mkdir -p "$appdir/usr/bin" "$appdir/usr/share/icons/hicolor/512x512/apps" "$appdir/usr/share/applications"
cp "$dir/quranku" "$appdir/usr/bin/quranku"
# Its resources sit next to the executable.
cp "$dir/icon.png" "$appdir/usr/bin/icon.png"
cp "$dir/quranku.desktop" "$appdir/quranku.desktop"
cp "$dir/quranku.desktop" "$appdir/usr/share/applications/quranku.desktop"
cp "$dir/quranku-512x512.png" "$appdir/quranku.png"
cp "$dir/quranku-512x512.png" "$appdir/usr/share/icons/hicolor/512x512/apps/quranku.png"
ln -s quranku.png "$appdir/.DirIcon"
cat > "$appdir/AppRun" <<'RUN'
#!/bin/sh
here=$(dirname "$(readlink -f "$0")")
exec "$here/usr/bin/quranku" "$@"
RUN
chmod +x "$appdir/AppRun"

out="$dir/Quranku-$version-$arch.AppImage"
ARCH=$arch appimagetool --no-appstream "$appdir" "$out"
echo "$out"
