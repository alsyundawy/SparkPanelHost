SHELL := /usr/bin/env bash
export PATH := $(shell go env GOPATH)/bin:$(PATH)
.DEFAULT_GOAL := help

APP_NAME      := sparkpanelhost
VERSION       ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
COMMIT_SHA    := $(shell git rev-parse --short HEAD 2>/dev/null || echo "local")
BUILD_TIME    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

BIN_DIR       := dist/bin
PKG_DIR       := dist/packages
GO_DIR        := backend
FE_DIR        := frontend

LDFLAGS       := -s -w \
                 -X main.Version=$(VERSION) \
                 -X main.Commit=$(COMMIT_SHA) \
                 -X main.BuildTime=$(BUILD_TIME)

.PHONY: help proto build-agent build-web build-frontend build-all \
        test security-audit update-recipes verify-recipes \
        package-tarball package-deb vm-up clean

help: ## Menampilkan daftar perintah Makefile yang tersedia
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

proto: ## Generate kode Go dan TypeScript dari api/proto/agent/v1/agent.proto
	@echo "==> Linting & Generating gRPC Protobuf stubs (v7)..."
	cd api/proto && buf lint && buf generate

build-agent: ## Compile static binary panel-agent (Privileged Daemon)
	@echo "==> Building static binary: panel-agent ($(VERSION))..."
	@mkdir -p $(BIN_DIR)
	cd $(GO_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -trimpath -ldflags="$(LDFLAGS)" -o ../$(BIN_DIR)/panel-agent ./cmd/panel-agent

build-web: ## Compile static binary panel-web (Unprivileged Control Plane)
	@echo "==> Building static binary: panel-web ($(VERSION))..."
	@mkdir -p $(BIN_DIR)
	cd $(GO_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -trimpath -ldflags="$(LDFLAGS)" -o ../$(BIN_DIR)/panel-web ./cmd/panel-web

build-frontend: ## Build aset produksi React 19 + TypeScript + Tailwind v4
	@echo "==> Building frontend bundle..."
	cd $(FE_DIR) && npm run typecheck && npm run build
	@mkdir -p dist/web
	@rm -rf dist/web/* && cp -r $(FE_DIR)/dist/* dist/web/

build-all: proto build-agent build-web build-frontend ## Build seluruh komponen (Agent, API, Frontend)
	@echo "==> Build selesai. Artefak tersedia di ./dist/"

test: ## Menjalankan seluruh unit test Go (dengan race detector) dan Frontend
	cd $(GO_DIR) && go test -race -v ./...
	cd $(FE_DIR) && npm run test -- --run

security-audit: ## Menjalankan SAST & SCA (govulncheck, gosec, trivy, npm audit)
	@echo "==> Memeriksa kerentanan dependensi Go (govulncheck)..."
	cd $(GO_DIR) && govulncheck ./...
	@echo "==> Menjalankan Static Analysis Security Testing (gosec)..."
	cd $(GO_DIR) && gosec -exclude-generated ./...
	@echo "==> Memeriksa kerentanan dependensi Frontend..."
	cd $(FE_DIR) && npm audit --audit-level=high

verify-recipes: ## Memvalidasi seluruh file YAML di recipes/ terhadap JSON Schema
	python3 scripts/verify_recipes.py

update-recipes: ## Mengecek rilis upstream terbaru dan memperbarui versi + SHA-256 manifes
	python3 scripts/update_manifests.py
	python3 scripts/verify_recipes.py

package-tarball: build-all verify-recipes ## Membungkus rilis produksi ke arsip tar.zst
	@mkdir -p $(PKG_DIR)
	tar --zstd -cf $(PKG_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.zst \
		-C dist bin web \
		-C .. recipes installer
	sha256sum $(PKG_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.zst > \
		$(PKG_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.zst.sha256
	@echo "==> Paket rilis siap di $(PKG_DIR)/"

package-deb: build-all verify-recipes ## Membungkus rilis produksi ke paket Debian/Ubuntu (.deb)
	@mkdir -p $(PKG_DIR)
	bash scripts/build_deb.sh --version $(VERSION) --output $(PKG_DIR)
	@echo "==> Paket Debian/Ubuntu (.deb) siap di $(PKG_DIR)/"

vm-up: ## Menjalankan VM pengujian Ubuntu 26.04 & Debian 13 via Vagrant
	cd deploy/vagrant && vagrant up

clean: ## Membersihkan seluruh hasil build sementara
	rm -rf dist/ $(FE_DIR)/dist/
