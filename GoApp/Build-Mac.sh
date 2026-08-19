#!/usr/bin/env bash
# Builds AtomicAWG.app for macOS: a single native Go binary (Fyne UI + the
# WireGuard/AmneziaWG engine linked in directly, no bundled second
# executable) packaged as a menu-bar-only ("LSUIElement") application.
set -euo pipefail

project_dir="$(cd "$(dirname "$0")" && pwd)"
version="3.5.0"
requested_arch="${1:-arm64}"

case "$requested_arch" in
  arm64) goarch=arm64 ;;
  x64|amd64) goarch=amd64 ;;
  *) echo "Usage: ./Build-Mac.sh [arm64|x64]" >&2; exit 2 ;;
esac

command -v go >/dev/null || { echo "Go toolchain was not found." >&2; exit 1; }

# Build and sign in a plain local temp directory rather than under
# project_dir/dist directly: when the project lives inside an iCloud
# Drive / File Provider synced folder, macOS re-tags directories with
# com.apple.FinderInfo mid-operation, which makes codesign refuse to seal
# the bundle ("resource fork ... not allowed"). Signing away from any
# synced location sidesteps the race; the finished, already-signed bundle
# is then simply copied into place afterwards.
work_dir="$(mktemp -d)"
trap 'rm -rf "$work_dir"' EXIT
app_dir="$work_dir/AtomicAWG.app"
contents_dir="$app_dir/Contents"
binary_path="$contents_dir/MacOS/AtomicAWG"

final_app_dir="$project_dir/dist/macos-$requested_arch/AtomicAWG.app"
mkdir -p "$contents_dir/MacOS" "$contents_dir/Resources"

echo "Building for darwin/$goarch..."
(
  cd "$project_dir"
  GOOS=darwin GOARCH="$goarch" CGO_ENABLED=1 go build \
    -trimpath \
    -ldflags "-X main.appVersion=$version" \
    -o "$binary_path" \
    .
)

cp "$project_dir/assets/AtomicAWG.icns" "$contents_dir/Resources/AtomicAWG.icns"
sed "s/@VERSION@/$version/g" "$project_dir/packaging/macos/Info.plist.template" > "$contents_dir/Info.plist"
chmod +x "$binary_path"

# Resource forks / quarantine xattrs (e.g. inherited by the .icns asset from
# wherever it was originally downloaded) make codesign refuse to seal the
# bundle ("resource fork ... not allowed"), so strip them first.
if command -v xattr >/dev/null; then
  xattr -cr "$app_dir"
fi

if command -v codesign >/dev/null; then
  # Ad-hoc signature: satisfies local Gatekeeper checks for a self-built app.
  # Distributing beyond this machine still requires a real Developer ID.
  codesign --force --deep --sign - "$app_dir"
fi

mkdir -p "$project_dir/dist"
if [[ -d "$final_app_dir" ]]; then
  rm -rf "$final_app_dir.old" 2>/dev/null || true
  mv "$final_app_dir" "$final_app_dir.old.$(date +%s)"
fi
mkdir -p "$(dirname "$final_app_dir")"
cp -R "$app_dir" "$final_app_dir"

archive="$project_dir/dist/AtomicAWG-macOS-$requested_arch.zip"
if command -v ditto >/dev/null; then
  ditto -c -k --sequesterRsrc --keepParent "$app_dir" "$archive"
else
  (cd "$work_dir" && zip -qry "$archive" AtomicAWG.app)
fi

echo "Done: $final_app_dir"
echo "Done: $archive"
