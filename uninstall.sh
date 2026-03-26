#!/bin/bash
# ============================================================
#  Proxy Bandwidth Saver — Uninstaller
#  Usage:  sudo bash uninstall.sh
# ============================================================
set -euo pipefail

APP_NAME="proxy-bandwidth-saver"
SERVICE_USER="proxybw"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${GREEN}[OK]${NC}    $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $1"; }

if [[ $EUID -ne 0 ]]; then
    echo -e "${RED}[ERROR]${NC} Chay voi quyen root: sudo bash uninstall.sh"
    exit 1
fi

echo ""
echo -e "${BOLD}Go cai dat ${APP_NAME}${NC}"
echo ""

# Stop & disable service
if systemctl is-active --quiet "$APP_NAME" 2>/dev/null; then
    systemctl stop "$APP_NAME"
    info "Dung service"
fi
if systemctl is-enabled --quiet "$APP_NAME" 2>/dev/null; then
    systemctl disable "$APP_NAME" >/dev/null 2>&1
    info "Vo hieu hoa service"
fi

# Remove service file
if [[ -f "/etc/systemd/system/${APP_NAME}.service" ]]; then
    rm -f "/etc/systemd/system/${APP_NAME}.service"
    systemctl daemon-reload
    info "Xoa systemd service"
fi

# Remove binary
if [[ -f "/usr/local/bin/${APP_NAME}" ]]; then
    rm -f "/usr/local/bin/${APP_NAME}"
    info "Xoa binary"
fi

# Ask about data
DATA_DIR="/var/lib/${APP_NAME}"
if [[ -d "$DATA_DIR" ]]; then
    echo ""
    read -p "  Xoa du lieu tai ${DATA_DIR}? (y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$DATA_DIR"
        info "Xoa du lieu"
    else
        warn "Giu lai du lieu tai ${DATA_DIR}"
    fi
fi

# Remove user
if id "$SERVICE_USER" &>/dev/null; then
    userdel "$SERVICE_USER" 2>/dev/null || true
    info "Xoa user ${SERVICE_USER}"
fi

echo ""
echo -e "${GREEN}${BOLD}Go cai dat hoan tat!${NC}"
echo ""
