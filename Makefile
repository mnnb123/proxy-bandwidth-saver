APP_NAME := proxy-bandwidth-saver
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

# Directories
FRONTEND_DIR := frontend
SERVER_DIR := cmd/server
DIST_DIR := dist

.PHONY: all clean frontend build-linux build-windows dev help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

all: build-linux ## Build for Linux (default)

# === Frontend ===

frontend: ## Build frontend
	cd $(FRONTEND_DIR) && npm ci && npm run build

# === Copy frontend dist into cmd/server for embed ===

embed-frontend: frontend
	rm -rf $(SERVER_DIR)/frontend
	cp -r $(FRONTEND_DIR)/dist $(SERVER_DIR)/frontend

# === Linux build (for Ubuntu VPS) ===

build-linux: embed-frontend ## Build headless server for Linux amd64
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 ./$(SERVER_DIR)
	@echo "Built: $(DIST_DIR)/$(APP_NAME)-linux-amd64"

build-linux-arm64: embed-frontend ## Build headless server for Linux arm64
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 ./$(SERVER_DIR)
	@echo "Built: $(DIST_DIR)/$(APP_NAME)-linux-arm64"

# === Windows build (headless, no Wails) ===

build-windows: embed-frontend ## Build headless server for Windows amd64
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe ./$(SERVER_DIR)
	@echo "Built: $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe"

# === Wails desktop app (Windows only) ===

build-wails: ## Build Wails desktop app (requires Wails CLI)
	wails build

dev: ## Run Wails dev mode
	wails dev

# === Package for deployment ===

package-linux: build-linux ## Package Linux binary + config files
	mkdir -p $(DIST_DIR)/package
	cp $(DIST_DIR)/$(APP_NAME)-linux-amd64 $(DIST_DIR)/package/$(APP_NAME)
	cp deploy/proxy-bandwidth-saver.service $(DIST_DIR)/package/
	cp deploy/install.sh $(DIST_DIR)/package/
	chmod +x $(DIST_DIR)/package/install.sh
	cd $(DIST_DIR) && tar -czf $(APP_NAME)-linux-amd64.tar.gz -C package .
	rm -rf $(DIST_DIR)/package
	@echo "Package: $(DIST_DIR)/$(APP_NAME)-linux-amd64.tar.gz"

# === Clean ===

clean: ## Clean build artifacts
	rm -rf $(DIST_DIR)
	rm -rf $(SERVER_DIR)/frontend
