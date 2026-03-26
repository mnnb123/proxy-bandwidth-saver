# Proxy Bandwidth Saver

Ung dung quan ly proxy thong minh giup tiet kiem bang thong residential. Tu dong phan loai traffic va dinh tuyen qua proxy phu hop (direct/datacenter/residential) dua tren rules do nguoi dung cau hinh.

## Tinh nang

- **Proxy Server** — HTTP + SOCKS5 proxy server tich hop
- **Traffic Rules** — Phan loai traffic theo domain, content-type, URL pattern
- **Smart Routing** — Tu dong dinh tuyen qua direct, datacenter hoac residential proxy
- **Caching** — Cache memory + disk de giam request trung lap
- **Budget Tracking** — Theo doi bang thong va chi phi theo ngay/thang
- **HTTPS Inspection** — MITM proxy cho HTTPS (tuy chon)
- **Proxy Authentication** — Bao mat bang username/password hoac IP whitelist
- **Port Mapping** — Moi upstream proxy duoc map thanh 1 output port rieng
- **Dark Mode** — Giao dien sang/toi voi chuyen doi muot ma
- **Responsive UI** — Sidebar tu dong thu gon khi cua so nho

## Cai dat len VPS (1 lenh)

SSH vao VPS va chay:

```bash
curl -sSL https://raw.githubusercontent.com/mnnb123/proxy-bandwidth-saver/master/setup.sh | sudo bash
```

Script se tu dong:
- Cai Go, Node.js, git (neu chua co)
- Clone repo, build frontend + backend
- Tao systemd service va khoi dong

Sau khi cai xong:
- **Web Panel**: `http://<vps-ip>:8080`
- **HTTP Proxy**: `<vps-ip>:8888`
- **SOCKS5 Proxy**: `<vps-ip>:8889`

> **Luu y**: Vao Settings > Proxy Authentication de bat bao mat truoc khi su dung!

Ho tro: Ubuntu/Debian, CentOS/RHEL, Fedora — amd64 & arm64.

---

<details>
<summary>Cai dat thu cong (advanced)</summary>

### Yeu cau he thong

**Desktop (Windows — Wails App)**
- Go >= 1.21, Node.js >= 18
- Wails CLI v2 — `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

**Headless Server (Linux VPS)**
- Go >= 1.21, Node.js >= 18

### Build & Deploy thu cong

```bash
# Clone
git clone https://github.com/mnnb123/proxy-bandwidth-saver.git
cd proxy-bandwidth-saver

# Dependencies
go mod download
cd frontend && npm install && cd ..

# Desktop (Windows)
wails dev          # dev mode
wails build        # build .exe

# Headless (Linux)
make build-linux        # amd64
make build-linux-arm64  # arm64

# Deploy
make package-linux
scp dist/proxy-bandwidth-saver-linux-amd64.tar.gz user@vps:/tmp/
# Tren VPS:
cd /tmp && tar xzf proxy-bandwidth-saver-linux-amd64.tar.gz
sudo bash install.sh
```

</details>

## Cau hinh

Tat ca cau hinh duoc thay doi qua giao dien Settings. Cac gia tri mac dinh:

| Setting | Mac dinh | Mo ta |
|---------|----------|-------|
| HTTP Port | `8888` | Port cho HTTP proxy |
| SOCKS5 Port | `8889` | Port cho SOCKS5 proxy |
| Bind Address | `127.0.0.1` | `0.0.0.0` de mo cho mang ngoai |
| Cache Memory | `512 MB` | Gioi han cache trong RAM |
| Cache Disk | `2048 MB` | Gioi han cache tren disk |
| MITM | `Off` | Bat de inspect HTTPS traffic |

### Bien moi truong

| Bien | Mo ta |
|------|-------|
| `PBS_DATA_DIR` | Thu muc luu tru data (DB, cache, certs) |

### Du lieu luu tru

| OS | Duong dan |
|----|-----------|
| Windows | `%APPDATA%\ProxyBandwidthSaver\` |
| Linux (user) | `~/.local/share/proxy-bandwidth-saver/` |
| Linux (service) | `/var/lib/proxy-bandwidth-saver/` |

## Bao mat Proxy

Khi bind `0.0.0.0` (mo cho mang ngoai), **bat buoc** phai bat 1 trong 2:

1. **Username/Password** — Proxy yeu cau xac thuc (HTTP Basic + SOCKS5 auth)
2. **IP Whitelist** — Chi cho phep cac IP/CIDR duoc cau hinh (localhost luon duoc phep)

Cau hinh trong Settings > Proxy Authentication.

## Quan ly Service (Linux)

```bash
# Xem trang thai
sudo systemctl status proxy-bandwidth-saver

# Dung / Khoi dong lai
sudo systemctl stop proxy-bandwidth-saver
sudo systemctl restart proxy-bandwidth-saver

# Xem log realtime
sudo journalctl -u proxy-bandwidth-saver -f
```

## Cau truc du an

```
proxy-bandwidth-saver/
├── app.go                  # Wails app (desktop mode)
├── main.go                 # Wails entry point
├── cmd/server/             # Headless server (VPS mode)
├── frontend/               # React + TypeScript + Vite
│   ├── src/
│   │   ├── components/     # UI components (Button, Modal, Sidebar...)
│   │   ├── pages/          # Dashboard, Rules, Proxies, Settings
│   │   ├── stores/         # Zustand state management
│   │   └── style.css       # Design system (CSS variables)
│   └── wailsjs/            # Auto-generated Wails bindings
├── internal/
│   ├── cache/              # Memory + disk cache
│   ├── classifier/         # Traffic classification rules
│   ├── config/             # App configuration
│   ├── database/           # SQLite database
│   ├── meter/              # Bandwidth metering
│   ├── optimizer/          # Request optimization
│   ├── proxy/              # HTTP/SOCKS5 proxy server, auth, MITM
│   ├── upstream/           # Upstream proxy manager & strategy
│   └── webapi/             # REST API + SSE (headless mode)
├── deploy/                 # Systemd service + install script
├── Makefile                # Build commands
└── wails.json              # Wails config
```

## Tech Stack

- **Backend**: Go, Wails v2, SQLite, bbolt
- **Frontend**: React 18, TypeScript, Vite, Tailwind CSS v4, Zustand, Recharts, Sonner
- **Proxy**: net/http, goproxy (MITM), go-socks5
- **Deploy**: Systemd, cross-compile (CGO_ENABLED=0)

## Make Commands

```bash
make help             # Xem tat ca commands
make dev              # Chay Wails dev mode
make build-linux      # Build headless cho Linux amd64
make build-windows    # Build headless cho Windows
make package-linux    # Dong goi de deploy
make clean            # Xoa build artifacts
```

## License

Private project.
