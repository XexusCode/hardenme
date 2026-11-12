#!/bin/sh
# hardenme installer for macOS.
#
#   curl -fsSL https://github.com/XexusCode/hardenme/releases/latest/download/install.sh | sh
#
# What it does, step by step (read it before running it, it is short):
#   1. Downloads the latest hardenme release from GitHub with curl.
#   2. Checks the download against the published SHA-256 checksum.
#   3. Installs the program in ~/.local/bin (no sudo, nothing outside your user folder).
#   4. Creates hardenme.app in Applications: double-click it to check your Mac
#      and choose what to fix.
# Uninstall:  curl -fsSL .../install.sh | sh -s -- --uninstall
#
# Options (environment variables):
#   HARDENME_VERSION   release tag to install, e.g. v0.1.0 (default: latest)
#   HARDENME_PREFIX    install prefix (default: ~/.local)
#   HARDENME_BASE_URL  where to download from (default: the GitHub release)
#   HARDENME_APP_DIR   where to create the app (default: /Applications or ~/Applications)
#   HARDENME_NO_OPEN=1 do not show the app in Finder at the end
set -eu

REPO="XexusCode/hardenme"
VERSION="${HARDENME_VERSION:-latest}"
PREFIX="${HARDENME_PREFIX:-$HOME/.local}"
BIN_DIR="$PREFIX/bin"
SHARE_DIR="$PREFIX/share/hardenme"
# /Applications is writable without sudo for admin users; otherwise use ~/Applications.
if [ -n "${HARDENME_APP_DIR:-}" ]; then
	APP_DIR="$HARDENME_APP_DIR"
elif [ -w /Applications ]; then
	APP_DIR="/Applications/hardenme.app"
else
	APP_DIR="$HOME/Applications/hardenme.app"
fi
ARCHIVE="hardenme_darwin_universal.tar.gz"

if [ -n "${HARDENME_BASE_URL:-}" ]; then
	BASE_URL="$HARDENME_BASE_URL"
elif [ "$VERSION" = "latest" ]; then
	BASE_URL="https://github.com/$REPO/releases/latest/download"
else
	BASE_URL="https://github.com/$REPO/releases/download/$VERSION"
fi

# Use the locale variables, or the macOS language list when they are unset
# (Terminal does not always set LANG).
LOCALE="${LC_ALL:-${LC_MESSAGES:-${LANG:-}}}"
case "$LOCALE" in
"" | C | POSIX) LOCALE="$(defaults read -g AppleLanguages 2>/dev/null | awk 'NR == 2 { gsub(/[ ",]/, ""); print }')" ;;
esac
case "$LOCALE" in
es*) LANG_ES=1 ;;
*) LANG_ES=0 ;;
esac

# say "english" "español"
say() {
	if [ "$LANG_ES" = 1 ]; then printf '%s\n' "$2"; else printf '%s\n' "$1"; fi
}

fail() {
	say "Error: $1" "Error: $2" >&2
	exit 1
}

uninstall() {
	rm -f "$BIN_DIR/hardenme"
	rm -rf "$SHARE_DIR" "$APP_DIR"
	say "hardenme has been removed. Your saved profile is in ~/.config/hardenme (delete it if you want)." \
		"hardenme se ha desinstalado. Tu perfil guardado está en ~/.config/hardenme (bórralo si quieres)."
}

if [ "${1:-}" = "--uninstall" ]; then
	uninstall
	exit 0
fi

[ "$(uname -s)" = "Darwin" ] || fail "for now hardenme only works on macOS." "por ahora hardenme solo funciona en macOS."
command -v curl >/dev/null 2>&1 || fail "curl is required." "hace falta curl."
command -v shasum >/dev/null 2>&1 || fail "shasum is required." "hace falta shasum."

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT INT TERM

say "1/4 Downloading hardenme..." "1/4 Descargando hardenme..."
curl -fsSL "$BASE_URL/$ARCHIVE" -o "$TMP/$ARCHIVE" || fail "could not download $BASE_URL/$ARCHIVE" "no se pudo descargar $BASE_URL/$ARCHIVE"
curl -fsSL "$BASE_URL/checksums.txt" -o "$TMP/checksums.txt" || fail "could not download the checksums." "no se pudieron descargar los checksums."

say "2/4 Checking that the download has not been tampered with..." "2/4 Comprobando que la descarga no se ha alterado..."
EXPECTED="$(awk -v f="$ARCHIVE" '$2 == f || $2 == "*" f { print $1 }' "$TMP/checksums.txt")"
[ -n "$EXPECTED" ] || fail "$ARCHIVE is not listed in checksums.txt." "$ARCHIVE no aparece en checksums.txt."
ACTUAL="$(shasum -a 256 "$TMP/$ARCHIVE" | awk '{ print $1 }')"
[ "$EXPECTED" = "$ACTUAL" ] || fail "checksum mismatch, the download may be corrupted. Nothing was installed." \
	"el checksum no coincide, la descarga puede estar dañada. No se ha instalado nada."

say "3/4 Installing in $BIN_DIR..." "3/4 Instalando en $BIN_DIR..."
tar -xzf "$TMP/$ARCHIVE" -C "$TMP"
[ -f "$TMP/hardenme" ] || fail "the archive does not contain hardenme." "el archivo no contiene hardenme."
mkdir -p "$BIN_DIR" "$SHARE_DIR"
install -m 0755 "$TMP/hardenme" "$BIN_DIR/hardenme"

# The script Terminal runs when you double-click the app: the scan, then
# the menu to fix what it found. hardenme picks the language from the
# system settings each time it runs.
if [ "$LANG_ES" = 1 ]; then
	CLOSE_MSG="Pulsa Intro para cerrar. "
else
	CLOSE_MSG="Press Return to close. "
fi
cat >"$SHARE_DIR/scan.command" <<EOF
#!/bin/sh
clear
HARDENME_NO_FIX_HINT=1 "$BIN_DIR/hardenme" scan
echo
"$BIN_DIR/hardenme" fix
echo
printf '%s' "$CLOSE_MSG"
read -r _
EOF
chmod 0755 "$SHARE_DIR/scan.command"

say "4/4 Creating $APP_DIR..." "4/4 Creando $APP_DIR..."
rm -rf "$APP_DIR"
mkdir -p "$APP_DIR/Contents/MacOS"
cat >"$APP_DIR/Contents/Info.plist" <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key><string>hardenme</string>
	<key>CFBundleDisplayName</key><string>hardenme</string>
	<key>CFBundleIdentifier</key><string>io.github.xexuscode.hardenme</string>
	<key>CFBundleExecutable</key><string>hardenme-launcher</string>
	<key>CFBundlePackageType</key><string>APPL</string>
	<key>CFBundleShortVersionString</key><string>1.0</string>
	<key>LSMinimumSystemVersion</key><string>12.0</string>
</dict>
</plist>
EOF
# Opening a .command file makes Terminal run it, without asking for
# permission to control Terminal.
cat >"$APP_DIR/Contents/MacOS/hardenme-launcher" <<EOF
#!/bin/sh
exec open -a Terminal "$SHARE_DIR/scan.command"
EOF
chmod 0755 "$APP_DIR/Contents/MacOS/hardenme-launcher"

echo
say "Done! hardenme $("$BIN_DIR/hardenme" --version | awk '{ print $NF }') is installed." \
	"¡Listo! hardenme $("$BIN_DIR/hardenme" --version | awk '{ print $NF }') está instalado."
say "To check your Mac, double-click hardenme in the Finder window that just opened (it is in $(dirname "$APP_DIR"))." \
	"Para revisar tu Mac, haz doble clic en hardenme en la ventana del Finder que se acaba de abrir (está en $(dirname "$APP_DIR"))."
case ":$PATH:" in
*":$BIN_DIR:"*) ;;
*) say "In Terminal you can also run: $BIN_DIR/hardenme scan" "En Terminal también puedes ejecutar: $BIN_DIR/hardenme scan" ;;
esac

# Show the app in Finder so the next step is a double-click.
if [ "${HARDENME_NO_OPEN:-0}" != 1 ]; then
	open -R "$APP_DIR" 2>/dev/null || true
fi
