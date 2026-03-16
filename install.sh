#!/bin/sh
set -e

REPO="nickw409/vex"
INSTALL_DIR="${VEX_INSTALL_DIR:-/usr/local/bin}"
VERSION="${VEX_VERSION:-latest}"

main() {
    os="$(detect_os)"
    arch="$(detect_arch)"

    if [ "$VERSION" = "latest" ]; then
        VERSION="$(fetch_latest_version)"
    fi

    if [ -z "$VERSION" ]; then
        err "could not determine latest version"
    fi

    # Strip leading v for filename
    vtag="$VERSION"
    vnum="${VERSION#v}"

    filename="vex_${vnum}_${os}_${arch}.tar.gz"
    url="https://github.com/${REPO}/releases/download/${vtag}/${filename}"

    tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT

    log "Downloading vex ${vtag} for ${os}/${arch}..."
    download "$url" "$tmpdir/$filename"

    log "Extracting..."
    tar -xzf "$tmpdir/$filename" -C "$tmpdir"

    log "Installing to ${INSTALL_DIR}/vex..."
    install_binary "$tmpdir/vex" "$INSTALL_DIR/vex"

    log "vex ${vtag} installed successfully"
}

detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        *)       err "unsupported OS: $(uname -s)" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *)             err "unsupported architecture: $(uname -m)" ;;
    esac
}

fetch_latest_version() {
    url="https://api.github.com/repos/${REPO}/releases/latest"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$url" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//'
    else
        err "curl or wget required"
    fi
}

download() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$2" "$1"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$2" "$1"
    else
        err "curl or wget required"
    fi
}

install_binary() {
    if [ -w "$INSTALL_DIR" ]; then
        cp "$1" "$2"
        chmod +x "$2"
    else
        log "Elevated permissions required for ${INSTALL_DIR}"
        sudo cp "$1" "$2"
        sudo chmod +x "$2"
    fi
}

log() {
    printf '%s\n' "$1" >&2
}

err() {
    log "error: $1"
    exit 1
}

main
