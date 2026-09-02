include makefile.project.mk

.DEFAULT_GOAL      := help

GOPATH             := $(shell go env GOPATH)
GOBIN              := $(shell go env GOBIN)
GOROOT             := $(shell go env GOROOT)
GOCACHE            := $(shell go env GOCACHE)
GOHOST_ARCH        := $(shell go env GOARCH)
GOHOST_OS          := $(shell go env GOOS)

EXE_EXT            := $(if $(filter windows,$(GOHOST_OS)),.exe,)

ifeq ($(strip $(GOPATH)),)
$(warning GOPATH is empty - is Go installed and on PATH? Go tools will not be found.)
endif

GOPATH_BIN         := $(GOPATH)/bin
GOROOT_BIN         := $(GOROOT)/bin

ifeq ($(strip $(GOBIN)),$(strip $(GOROOT_BIN)))
GO_BIN_DIR        := $(GOPATH_BIN)
else
GO_BIN_DIR        := $(if $(strip $(GOBIN)),$(GOBIN),$(GOPATH_BIN))
endif

HOME_LOCAL_BIN     := $(HOME)/.local/bin

PROGRAMS_POSIX     := $(subst \,/,$(PROGRAMS))
MSYS64_USR_BIN     := $(PROGRAMS_POSIX)/msys64/usr/bin
MSYS64_UCRT64_BIN  := $(PROGRAMS_POSIX)/msys64/ucrt64/bin

empty             :=
space             := $(empty) $(empty)

posix-path         = $(if $(strip $(1)),$(shell \
                        d='$(subst \,/,$(strip $(1)))'; [ -d "$$d" ] || exit 0; \
                        command -v cygpath >/dev/null 2>&1 && cygpath -u "$$d" 2>/dev/null || printf '%s' "$$d"))

uniq               = $(if $(1),$(firstword $(1)) $(call uniq,$(filter-out $(firstword $(1)),$(1))))

join-path          = $(subst $(space),:,$(strip $(1)))

TOOL_DIRS          := \
    $(call posix-path,$(GOBIN)) \
    $(call posix-path,$(GOPATH_BIN)) \
    $(call posix-path,$(GOROOT_BIN)) \
    $(call posix-path,$(HOME_LOCAL_BIN)) \
    $(call posix-path,$(MSYS64_UCRT64_BIN)) \
    $(call posix-path,$(MSYS64_USR_BIN))

PATH_PREPEND       := $(call join-path,$(call uniq,$(TOOL_DIRS)))

export PATH        := $(PATH_PREPEND):$(PATH)

GO_BIN_DIR_POSIX   := $(call posix-path,$(GO_BIN_DIR))
GO_BIN_DIR_PATH    := $(GO_BIN_DIR_POSIX)

MKDIR              := mkdir
AWK                := awk
FIND               := find

GOBUILD            := go build -a

CYCLONEDX          := cyclonedx-gomod
DEADCODE           := deadcode
GOLANGCI_LINT      := golangci-lint
GORELEASER         := goreleaser
GOVULNCHECK        := govulncheck
WINRES             := go-winres

ifndef CGO_ENABLED
CGO_ENABLED := 0
endif

.PHONY: help
help: ## Show this help
	@echo "Usage: make [target]"
	@grep -hE '^[a-zA-Z0-9_.-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| $(AWK) 'BEGIN {FS = ":.*?## "}; {printf "  %-28s %s\n", $$1, $$2}' \
		| sort

.PHONY: clean
clean: go-winres_clean ## Remove generated build artifacts
	@echo "Running clean..."
	rm -rf $(DIST_DIR)
	@echo "clean: done"

.PHONY: go-install_development_tools
go-install_development_tools: ## Install required development tools
	@echo "Running install for required developer tools..."
	go install golang.org/x/tools/cmd/deadcode@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install github.com/tc-hib/go-winres@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/goreleaser/goreleaser/v2@latest
	go install github.com/anchore/syft/cmd/syft@latest
	go install github.com/anchore/grype/cmd/grype@latest
	go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest
	@echo "install-tools: deadcode, golangci-lint, go-winres, govulncheck, goreleaser, syft*, grype*, cyclonedx-gomod*, installed"

.PHONY: build-windows
build-windows: go-winres ## Build for windows-amd64
	@echo "Running Windows amd64 build..."
	@mkdir -p $(DIST_DIR)/windows-amd64
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/windows-amd64/$(BINARY)$(EXE_EXT) $(MAIN)
	@env GOOS=windows GOARCH=amd64 go env GOOS GOARCH GOROOT GOPATH

.PHONY: build-linux
build-linux: go-vet ## Build for linux-amd64
	@echo "Running Linux amd64 build..."
	@mkdir -p $(DIST_DIR)/linux-amd64
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/linux-amd64/$(BINARY) $(MAIN)
	@env GOOS=linux GOARCH=amd64 go env GOOS GOARCH GOROOT GOPATH

DEADCODE_STRICT ?=

.PHONY: go-deadcode
go-deadcode: ## Run deadcode analysis
	@echo "Running deadcode analysis..."
	@if ! command -v $(DEADCODE) >/dev/null 2>&1; then \
		echo "ERROR: deadcode not found - run: make dev-install_tools"; exit 1; \
	elif $(DEADCODE) -test ./...; then \
		: ; \
	elif [ -n "$(DEADCODE_STRICT)" ]; then \
		echo "ERROR: deadcode failed"; exit 1; \
	else \
		echo "WARNING: deadcode failed - skipping (set DEADCODE_STRICT=1 to make this fatal)"; \
	fi

.PHONY: go-tidy
go-tidy: go-deadcode ## Run go mod tidy
	@echo "Running go mod tidy..."
	go mod tidy

.PHONY: go-verify
go-verify: go-tidy ## Run go mod verify
	@echo "Running go mod verify..."
	go mod verify

.PHONY: go-gen
go-gen: go-tidy ## Run go generate
	@echo "Running go generate..."
	go generate

.PHONY: go-fmt
go-fmt: go-gen ## Run go fmt
	@echo "Running go fmt..."
	go fmt ./...

.PHONY: go-lint
go-lint: go-fmt ## Run golangci-lint
	@echo "Running golangci-lint..."
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then $(GOLANGCI_LINT) run --timeout=5m ./...; \
	else echo "ERROR: golangci-lint not found - run: make dev-install_tools"; exit 1; fi

.PHONY: go-vet
go-vet: go-lint ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: go-winres
go-winres: go-vet go-winres_clean ## Generate Windows resource files
	@echo "Running go-winres, resource files (.syso) generation..."
	@command -v $(WINRES) >/dev/null 2>&1 || { echo "go-winres not found - run: make dev-install_tools"; exit 1; }
	@[ -f winres/winres.json ] || { echo "winres/winres.json not found - initialising scaffold..."; mkdir -p winres; $(WINRES) init; }
	@$(WINRES) make --arch amd64 --product-version "$(VERSION)" --file-version "$(VERSION)" --in winres/winres.json
	@echo "go-winres: .syso files generated"

.PHONY: go-winres_clean
go-winres_clean: ## Remove generated .syso resource files
	@echo "Running go-winres_clean, resource files (.syso) cleanup..."
	$(FIND) . -name "*.syso" -delete
	@echo "go-winres_clean: done"

.PHONY: go-clean_cache
go-clean_cache: ## Run go clean --cache
	@echo "Running Go build cache cleanup..."
	go clean --cache

.PHONY: go-releaser_snapshot
go-releaser_snapshot: ## Run goreleaser snapshot
	@echo "Running goreleaser snapshot..."
	$(GORELEASER) release --snapshot --clean

.PHONY: go-releaser_tag_local
go-releaser_tag_local: ## Run goreleaser release from tag (local, no publish)
	@echo "Running goreleaser release (local, no publish)..."
	$(GORELEASER) release --clean --skip=publish

.PHONY: go-releaser_check
go-releaser_check: ## Validate goreleaser config
	@echo "Running goreleaser config validation..."
	$(GORELEASER) check

.PHONY: go-vulncheck
go-vulncheck: ## Run go-vulncheck
	@echo "Running go-vulncheck..."
	$(GOVULNCHECK) ./...

.PHONY: go-vulncheck_report
go-vulncheck_report: ## Run go-vulncheck_report
	@echo "Running go-vulncheck_report..."
	@$(MKDIR) -p $(DIST_DIR)
	@echo "# Vulnerability Report" > $(DIST_DIR)/vuln-check.md
	@echo "Generated on: $$(date)" >> $(DIST_DIR)/vuln-check.md
	@echo "" >> $(DIST_DIR)/vuln-check.md
	$(GOVULNCHECK) -show verbose -format=text ./... >> $(DIST_DIR)/vuln-check.md
	$(GOVULNCHECK) -format=sarif ./... > $(DIST_DIR)/vuln-check.sarif
	@echo "Vulnerability check generated in $(DIST_DIR)/ directory"
