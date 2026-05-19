# ═══════════════════════════════════════════════════════
# Blogo — Monorepo 工程化 Makefile (Windows / Linux)
# ═══════════════════════════════════════════════════════

.PHONY: dev server web admin build build-server build-web build-admin docker clean help

SERVER_DIR  := blogo-server
WEB_DIR     := blogo-web
ADMIN_DIR   := blogo-admin
DOCKER_DIR  := deploy/docker

# 根目录检测
ROOT_MARKER := $(wildcard $(SERVER_DIR)/main.go)
ifeq ($(ROOT_MARKER),)
  $(error Please run from project root: Blogo/)
endif

# OS 检测
ifneq ($(OS),Windows_NT)
  NPM := npm run
else
  NPM := cmd //c npm run
endif

# ═══════════════════════════════════════════════════════
# 开发服务
# ═══════════════════════════════════════════════════════

server:
	cd $(SERVER_DIR) && air

web:
	cd $(WEB_DIR) && $(NPM) dev

admin:
	cd $(ADMIN_DIR) && $(NPM) dev

# ═══════════════════════════════════════════════════════
# 构建
# ═══════════════════════════════════════════════════════

build: build-server build-web build-admin

build-server:
	cd $(SERVER_DIR) && $(MAKE) build

build-web:
	cd $(WEB_DIR) && npm run build

build-admin:
	cd $(ADMIN_DIR) && npm run build

# ═══════════════════════════════════════════════════════
# Docker
# ═══════════════════════════════════════════════════════

docker:
	docker build -t blogo-server:latest -f $(DOCKER_DIR)/Dockerfile $(SERVER_DIR)

# ═══════════════════════════════════════════════════════
# 代码生成
# ═══════════════════════════════════════════════════════

wire:
	cd $(SERVER_DIR) && $(MAKE) wire

swagger:
	cd $(SERVER_DIR) && $(MAKE) swagger

# ═══════════════════════════════════════════════════════
# 清理
# ═══════════════════════════════════════════════════════

clean:
	cd $(SERVER_DIR) && $(MAKE) clean
ifneq ($(OS),Windows_NT)
	rm -rf $(WEB_DIR)/dist
	rm -rf $(ADMIN_DIR)/dist
	find logs -name '*.log' -delete 2>/dev/null || true
	rm -rf storage/temp/*
else
	cmd //c "if exist $(WEB_DIR)\dist rmdir /s /q $(WEB_DIR)\dist"
	cmd //c "if exist $(ADMIN_DIR)\dist rmdir /s /q $(ADMIN_DIR)\dist"
	cmd //c "del /s /q logs\*.log 2>nul"
	cmd //c "del /q storage\temp\*.log 2>nul"
endif

# ═══════════════════════════════════════════════════════
# 帮助
# ═══════════════════════════════════════════════════════

help:
	@echo "Blogo Makefile Commands:"
	@echo ""
	@echo "  Dev:"
	@echo "    make server        Start Go backend (Air)"
	@echo "    make web           Start web frontend (Vite)"
	@echo "    make admin         Start admin panel (Vite)"
	@echo ""
	@echo "  Build:"
	@echo "    make build         Build all"
	@echo "    make build-server  Build Go backend"
	@echo "    make build-web     Build web frontend"
	@echo "    make build-admin   Build admin panel"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker        Build Docker image"
	@echo ""
	@echo "  Code:"
	@echo "    make wire          Generate Wire DI code"
	@echo "    make swagger       Generate Swagger docs"
	@echo ""
	@echo "  Tools:"
	@echo "    make clean         Clean build artifacts"
	@echo "    make help          Show this help"
