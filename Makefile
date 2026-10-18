VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

PKG      := github.com/QuietSugar/hall-kit/internal/version
LDFLAGS  := -X $(PKG).Version=$(VERSION) -X $(PKG).Commit=$(COMMIT) -X $(PKG).Date=$(DATE)

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build hall into bin/ and create the short names there
	go build -ldflags '$(LDFLAGS)' -o bin/hall ./cmd/hall
	./bin/hall aliases

.PHONY: install
install: ## go install hall and create the short names in GOBIN
	go install -ldflags '$(LDFLAGS)' ./cmd/hall
	@d=$$(go env GOBIN); [ -n "$$d" ] || d=$$(go env GOPATH)/bin; "$$d/hall" aliases

.PHONY: test
test: ## Run all tests
	go test ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: fmt
fmt: ## Format all Go files with gofmt
	gofmt -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: clean
clean: ## Remove the bin/ directory
	rm -rf bin/
