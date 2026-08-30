# ── Project settings ─────────────────────────────────────────────────────────

BINARY             := m2h
PRODUCTNAME        := m2h
VERSION            := 1.0.3

DIST_DIR           := ./dist

MAIN               := .
MAIN_PKG           := .

LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: go-install_project_modules
go-install_project_modules: ## Install project specific required Go modules
	@echo "Running install for required Go modules..."
	go install github.com/spf13/cobra/cobra-cli@latest