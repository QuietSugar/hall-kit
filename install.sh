#!/bin/sh
# hall-kit installer: download a prebuilt release and install the binaries.
#
#   curl -fsSL https://raw.githubusercontent.com/QuietSugar/hall-kit/master/install.sh | sh
#
# The binary lands in ~/.hall/bin, which the shell block added by
# "hall shell install" puts on PATH together with the short tool names.
#
# 访问不了 GitHub 时，用加速前缀拉这个脚本即可（例如
# https://gh-proxy.com/https://raw.githubusercontent.com/.../install.sh）；
# 脚本下载资产时会自动改用同一个加速前缀，GH_PROXY 可覆盖它。
# 标准代理 https_proxy 依然生效，且会同时作用于拉脚本与下载资产。
#
# Env overrides:
#   HALL_KIT_VERSION      release tag (default latest), e.g. v0.0.1
#   HALL_KIT_INSTALL_DIR  target directory (default $HOME/.hall/bin,
#                         which the shell integration adds to PATH)
#   GH_PROXY              overrides the built-in GitHub mirror prefix
#   HALL_KIT_NO_ALIASES=1 skip creating the short tool names
set -eu

REPO="QuietSugar/hall-kit"
INSTALL_DIR="${HALL_KIT_INSTALL_DIR:-$HOME/.hall/bin}"
DEFAULT_GH_PROXY="https://gh-proxy.com/"

# hall_mirror <url> prints the mirrored URL; other URLs pass through.
hall_mirror() {
    case "$1" in
        https://raw.githubusercontent.com/*|https://github.com/*|https://codeload.github.com/*|https://api.github.com/*)
            m="${GH_PROXY:-$DEFAULT_GH_PROXY}"
            printf "%s%s\n" "${m%/}/" "$1" ;;
        *)
            printf "%s\n" "$1" ;;
    esac
}

# hall_download <url> [dest] downloads with curl; no dest writes to stdout.
hall_download() {
    if [ -n "${2:-}" ]; then
        curl -fsSL -o "$2" "$1"
    else
        curl -fsSL "$1"
    fi
}

# gh_fetch <url> [dest] honours GH_PROXY; otherwise it tries direct first and
# retries through the built-in mirror. The checksum only protects against
# transmission damage, not against a mirror serving a matching pair.
gh_fetch() {
    if [ -n "${GH_PROXY:-}" ]; then
        hall_download "$(hall_mirror "$1")" "${2:-}"
        return $?
    fi
    if hall_download "$1" "${2:-}"; then
        return 0
    fi
    echo "note: direct download failed, retrying via $DEFAULT_GH_PROXY" >&2
    hall_download "$(hall_mirror "$1")" "${2:-}"
}

os="$(uname -s)"
case "$os" in
    Linux) os=linux ;;
    Darwin) os=darwin ;;
    MINGW*|MSYS*|CYGWIN*) os=windows ;;
    *) echo "error: unsupported OS: $os (linux/darwin/windows-gitbash only)" >&2; exit 1 ;;
esac

arch="$(uname -m)"
case "$arch" in
    x86_64|amd64) arch=amd64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) echo "error: unsupported architecture: $arch" >&2; exit 1 ;;
esac

ext=""
if [ "$os" = windows ]; then ext=".exe"; fi

# Resolve the latest tag via the GitHub API (the mirror proxies api.github.com);
# fall back to the releases/latest redirect for the anonymous API rate limit.
tag="${HALL_KIT_VERSION:-latest}"
if [ "$tag" = latest ]; then
    tag="$(gh_fetch "https://api.github.com/repos/$REPO/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)" || true
fi
if [ -z "$tag" ] || [ "$tag" = latest ]; then
    tag="$(curl -fsSLI -o /dev/null -w "%{url_effective}" "https://github.com/$REPO/releases/latest" | sed "s#.*/tag/##")" || true
fi
case "$tag" in
    v*) ;;
    *) echo "error: HALL_KIT_VERSION must be a release tag like v0.0.1 (got: $tag)." >&2
       echo "       for branch/commit installs use: go install github.com/$REPO/cmd/hall@$tag" >&2
       exit 1 ;;
esac

archive="hall-kit_${tag}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/$tag"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT INT TERM

# A Windows curl (Cygwin resolves to C:\Windows\System32\curl.exe) does not
# understand POSIX paths for -o, so hand it a native path when cygpath exists.
dl="$tmp"
if [ "$os" = windows ] && command -v cygpath >/dev/null 2>&1; then
    dl="$(cygpath -w "$tmp")"
fi

echo "Downloading hall-kit $tag ($os/$arch)..."
gh_fetch "$base/$archive" "$dl/$archive"
gh_fetch "$base/checksums.txt" "$dl/checksums.txt"

expected="$(grep " $archive\$" "$tmp/checksums.txt" || true)"
if [ -z "$expected" ]; then
    echo "error: $archive is not listed in checksums.txt" >&2
    exit 1
fi
if command -v sha256sum >/dev/null 2>&1; then
    printf "%s\n" "$expected" | (cd "$tmp" && sha256sum -c - >/dev/null)
elif command -v shasum >/dev/null 2>&1; then
    printf "%s\n" "$expected" | (cd "$tmp" && shasum -a 256 -c - >/dev/null)
else
    echo "warning: sha256sum/shasum not found, skipping checksum verification" >&2
fi
echo "Checksum OK."

mkdir -p "$tmp/extract"
tar -xzf "$tmp/$archive" -C "$tmp/extract"
mkdir -p "$INSTALL_DIR"
cp -f "$tmp/extract/hall$ext" "$INSTALL_DIR/hall$ext"
chmod 0755 "$INSTALL_DIR/hall$ext"
echo "Installed $INSTALL_DIR/hall$ext"

case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *) echo ""
       echo "note: $INSTALL_DIR is not on your PATH. Add it in your shell rc, for example:"
       echo "  export PATH=\"$INSTALL_DIR:\$PATH\"" ;;
esac

if [ "$os" = windows ]; then
    echo "note: Windows support is experimental; run these binaries from git-bash."
fi

if [ "${HALL_KIT_NO_ALIASES:-}" = 1 ]; then
    echo "note: skipping short-name aliases (HALL_KIT_NO_ALIASES=1)"
else
    if "$INSTALL_DIR/hall$ext" aliases; then
        :
    else
        echo "warning: alias creation failed; run \"hall aliases\" manually." >&2
    fi
fi

if "$INSTALL_DIR/hall$ext" shell install; then
    :
else
    echo "warning: hall shell install failed; fix the cause and run it manually." >&2
fi

"$INSTALL_DIR/hall$ext" --version