#!/bin/bash
# ============================================================
#  Proxy Bandwidth Saver — One-click VPS Installer
#  Usage:  curl -sSL https://raw.githubusercontent.com/mnnb123/proxy-bandwidth-saver/master/setup.sh | sudo bash
# ============================================================
set -euo pipefail

APP_NAME="proxy-bandwidth-saver"
REPO_URL="https://github.com/mnnb123/proxy-bandwidth-saver.git"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/var/lib/${APP_NAME}"
SERVICE_USER="proxybw"
BUILD_DIR="/tmp/${APP_NAME}-build"

# Node.js version for frontend build
NODE_MAJOR=20

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${GREEN}[OK]${NC}    $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }
step()  { echo -e "\n${CYAN}${BOLD}>>> $1${NC}"; }

# ── Pre-checks ──────────────────────────────────────────────

if [[ $EUID -ne 0 ]]; then
    error "Chay voi quyen root:  curl -sSL <url> | sudo bash"
fi

ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  GOARCH="amd64" ;;
    aarch64) GOARCH="arm64" ;;
    *)       error "Architecture khong ho tro: $ARCH (chi ho tro amd64/arm64)" ;;
esac

echo ""
echo -e "${BOLD}========================================${NC}"
echo -e "${BOLD}  Proxy Bandwidth Saver — Installer${NC}"
echo -e "${BOLD}========================================${NC}"
echo -e "  Arch:   ${ARCH} (${GOARCH})"
echo -e "  OS:     $(. /etc/os-release 2>/dev/null && echo "$PRETTY_NAME" || uname -s)"
echo ""

# ── 1. Install Go ───────────────────────────────────────────

install_go() {
    step "1/6  Cai dat Go..."

    if command -v go &>/dev/null; then
        GO_VER=$(go version | awk '{print $3}' | sed 's/go//')
        GO_MAJOR=$(echo "$GO_VER" | cut -d. -f1)
        GO_MINOR=$(echo "$GO_VER" | cut -d. -f2)
        if [[ "$GO_MAJOR" -ge 1 && "$GO_MINOR" -ge 21 ]]; then
            info "Go ${GO_VER} da co san"
            return
        fi
        warn "Go ${GO_VER} qua cu, can >= 1.21. Cai phien ban moi..."
    fi

    GO_VERSION="1.23.4"
    GO_TAR="go${GO_VERSION}.linux-${GOARCH}.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TAR}"

    echo "  Downloading Go ${GO_VERSION}..."
    curl -sSL "$GO_URL" -o "/tmp/${GO_TAR}"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "/tmp/${GO_TAR}"
    rm -f "/tmp/${GO_TAR}"

    export PATH="/usr/local/go/bin:$PATH"
    export GOPATH="/root/go"
    export PATH="$GOPATH/bin:$PATH"

    info "Go $(go version | awk '{print $3}') da cai dat"
}

# ── 2. Install Node.js ─────────────────────────────────────

install_node() {
    step "2/6  Cai dat Node.js..."

    if command -v node &>/dev/null; then
        NODE_VER=$(node -v | sed 's/v//')
        NODE_M=$(echo "$NODE_VER" | cut -d. -f1)
        if [[ "$NODE_M" -ge 18 ]]; then
            info "Node.js v${NODE_VER} da co san"
            return
        fi
        warn "Node.js v${NODE_VER} qua cu, can >= 18. Cai phien ban moi..."
    fi

    if command -v apt-get &>/dev/null; then
        # Debian/Ubuntu
        apt-get update -qq
        apt-get install -y -qq ca-certificates curl gnupg >/dev/null 2>&1
        mkdir -p /etc/apt/keyrings
        curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg 2>/dev/null
        echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_${NODE_MAJOR}.x nodistro main" > /etc/apt/sources.list.d/nodesource.list
        apt-get update -qq
        apt-get install -y -qq nodejs >/dev/null 2>&1
    elif command -v dnf &>/dev/null; then
        # RHEL/Fedora
        curl -fsSL "https://rpm.nodesource.com/setup_${NODE_MAJOR}.x" | bash - >/dev/null 2>&1
        dnf install -y nodejs >/dev/null 2>&1
    elif command -v yum &>/dev/null; then
        # CentOS
        curl -fsSL "https://rpm.nodesource.com/setup_${NODE_MAJOR}.x" | bash - >/dev/null 2>&1
        yum install -y nodejs >/dev/null 2>&1
    else
        error "Khong the cai Node.js tu dong. Hay cai Node.js >= 18 thu cong."
    fi

    info "Node.js $(node -v) da cai dat"
}

# ── 3. Install build tools ──────────────────────────────────

install_deps() {
    step "3/6  Cai dat build tools..."

    if command -v apt-get &>/dev/null; then
        apt-get install -y -qq git make >/dev/null 2>&1
    elif command -v dnf &>/dev/null; then
        dnf install -y git make >/dev/null 2>&1
    elif command -v yum &>/dev/null; then
        yum install -y git make >/dev/null 2>&1
    fi

    info "git $(git --version | awk '{print $3}'), make OK"
}

# ── 4. Clone & Build ────────────────────────────────────────

build_app() {
    step "4/6  Clone repo & build..."

    rm -rf "$BUILD_DIR"
    git clone --depth 1 "$REPO_URL" "$BUILD_DIR" 2>&1 | tail -1
    cd "$BUILD_DIR"

    info "Building frontend..."
    cd frontend && npm ci --silent 2>&1 | tail -3
    npx vite build
    cd ..

    info "Embedding frontend..."
    rm -rf cmd/server/frontend
    cp -r frontend/dist cmd/server/frontend

    info "Building Go binary (linux/${GOARCH})..."
    export PATH="/usr/local/go/bin:/root/go/bin:$PATH"
    CGO_ENABLED=0 GOOS=linux GOARCH="$GOARCH" go build \
        -ldflags "-s -w" \
        -o "${APP_NAME}" \
        ./cmd/server

    info "Build thanh cong! ($(du -h ${APP_NAME} | awk '{print $1}'))"
}

# ── 5. Install ──────────────────────────────────────────────

install_app() {
    step "5/6  Cai dat vao he thong..."

    # Create service user
    if ! id "$SERVICE_USER" &>/dev/null; then
        useradd --system --no-create-home --shell /usr/sbin/nologin "$SERVICE_USER"
        info "Tao user: ${SERVICE_USER}"
    fi

    # Create data directory
    mkdir -p "$DATA_DIR"/{cache,certs}
    chown -R "$SERVICE_USER":"$SERVICE_USER" "$DATA_DIR"
    chmod 750 "$DATA_DIR"
    info "Data dir: ${DATA_DIR}"

    # Stop existing service
    if systemctl is-active --quiet "$APP_NAME" 2>/dev/null; then
        systemctl stop "$APP_NAME"
        info "Dung service cu"
    fi

    # Copy binary
    cp "${BUILD_DIR}/${APP_NAME}" "${INSTALL_DIR}/${APP_NAME}"
    chmod 755 "${INSTALL_DIR}/${APP_NAME}"
    info "Binary: ${INSTALL_DIR}/${APP_NAME}"

    # Install systemd service
    cp "${BUILD_DIR}/deploy/proxy-bandwidth-saver.service" /etc/systemd/system/
    systemctl daemon-reload
    info "Systemd service da cai dat"
}

# ── 6. Start ────────────────────────────────────────────────

start_app() {
    step "6/6  Khoi dong service..."

    systemctl enable "$APP_NAME" >/dev/null 2>&1
    systemctl start "$APP_NAME"

    sleep 2

    # Cleanup build dir
    rm -rf "$BUILD_DIR"

    VPS_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "<vps-ip>")

    if systemctl is-active --quiet "$APP_NAME"; then
        echo ""
        echo -e "${GREEN}${BOLD}========================================${NC}"
        echo -e "${GREEN}${BOLD}  Cai dat thanh cong!${NC}"
        echo -e "${GREEN}${BOLD}========================================${NC}"
        echo ""
        echo -e "  ${BOLD}Web Panel:${NC}    http://${VPS_IP}:8080"
        echo -e "  ${BOLD}HTTP Proxy:${NC}   ${VPS_IP}:8888"
        echo -e "  ${BOLD}SOCKS5 Proxy:${NC} ${VPS_IP}:8889"
        echo ""
        echo -e "  ${BOLD}Quan ly:${NC}"
        echo "    systemctl status  ${APP_NAME}"
        echo "    systemctl restart ${APP_NAME}"
        echo "    journalctl -u ${APP_NAME} -f"
        echo ""
        echo -e "  ${YELLOW}Luu y: Vao Settings > Proxy Authentication de bat bao mat!${NC}"
        echo ""
    else
        echo ""
        warn "Service khong khoi dong duoc. Kiem tra log:"
        echo "    journalctl -u ${APP_NAME} --no-pager -n 30"
        echo ""
    fi
}

# ── Main ────────────────────────────────────────────────────

install_go
install_node
install_deps
build_app
install_app
start_app
