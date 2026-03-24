#!/bin/bash
set -euo pipefail

APP_NAME="proxy-bandwidth-saver"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/var/lib/${APP_NAME}"
SERVICE_USER="proxybw"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Check root
if [[ $EUID -ne 0 ]]; then
    error "Script phải chạy với quyền root. Dùng: sudo bash install.sh"
fi

# Check binary exists
BINARY="./${APP_NAME}"
if [[ ! -f "$BINARY" ]]; then
    BINARY="./${APP_NAME}-linux-amd64"
fi
if [[ ! -f "$BINARY" ]]; then
    error "Không tìm thấy binary '${APP_NAME}'. Đảm bảo file binary nằm cùng thư mục."
fi

info "=== Cài đặt ${APP_NAME} ==="

# Create service user
if ! id "$SERVICE_USER" &>/dev/null; then
    info "Tạo user: ${SERVICE_USER}"
    useradd --system --no-create-home --shell /usr/sbin/nologin "$SERVICE_USER"
else
    info "User ${SERVICE_USER} đã tồn tại"
fi

# Create data directory
info "Tạo thư mục data: ${DATA_DIR}"
mkdir -p "$DATA_DIR"/{cache,certs}
chown -R "$SERVICE_USER":"$SERVICE_USER" "$DATA_DIR"
chmod 750 "$DATA_DIR"

# Stop existing service if running
if systemctl is-active --quiet "$APP_NAME" 2>/dev/null; then
    info "Dừng service đang chạy..."
    systemctl stop "$APP_NAME"
fi

# Install binary
info "Cài đặt binary vào ${INSTALL_DIR}/"
cp "$BINARY" "${INSTALL_DIR}/${APP_NAME}"
chmod 755 "${INSTALL_DIR}/${APP_NAME}"

# Install systemd service
SERVICE_FILE="${APP_NAME}.service"
if [[ -f "$SERVICE_FILE" ]]; then
    info "Cài đặt systemd service..."
    cp "$SERVICE_FILE" /etc/systemd/system/
    systemctl daemon-reload
else
    warn "Không tìm thấy ${SERVICE_FILE}, bỏ qua cài đặt service"
fi

# Enable and start
info "Kích hoạt và khởi động service..."
systemctl enable "$APP_NAME"
systemctl start "$APP_NAME"

# Wait a bit and check status
sleep 2
if systemctl is-active --quiet "$APP_NAME"; then
    info "=== Cài đặt thành công! ==="
    echo ""
    echo "  Web Admin Panel: http://$(hostname -I | awk '{print $1}'):8080"
    echo "  HTTP Proxy:      $(hostname -I | awk '{print $1}'):8888"
    echo "  SOCKS5 Proxy:    $(hostname -I | awk '{print $1}'):8889"
    echo ""
    echo "  Quản lý service:"
    echo "    sudo systemctl status ${APP_NAME}"
    echo "    sudo systemctl stop ${APP_NAME}"
    echo "    sudo systemctl restart ${APP_NAME}"
    echo "    sudo journalctl -u ${APP_NAME} -f"
    echo ""
else
    warn "Service không khởi động được. Kiểm tra log:"
    echo "    sudo journalctl -u ${APP_NAME} --no-pager -n 20"
fi
