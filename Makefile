.PHONY: help build run test test-fast test-integration lint fmt fmt-check ci clean cover cover-check

# The race detector needs cgo; some machines default it off.
RACE_ENV := CGO_ENABLED=1

# The default target: list the targets with their one-line descriptions
# (the `## ` comment after each target name).
.DEFAULT_GOAL := help
help: ## Show this list of targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  \033[36m%-17s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build bin/lazybrew
	go build -o bin/lazybrew ./cmd/lazybrew

run: build ## Build, then launch the TUI
	./bin/lazybrew

# The release/lane gate's test step.
test: ## Run the unit tests with the race detector
	$(RACE_ENV) go test ./... -race -count=1

# Quick per-commit check: no race detector. Pass PKG to narrow it,
# e.g. `make test-fast PKG=./internal/gui/...`.
PKG ?= ./...
test-fast: ## Run the unit tests without -race (PKG=./internal/gui/... narrows it)
	go test $(PKG) -count=1

test-integration: ## Run the integration tests (needs real Homebrew)
	$(RACE_ENV) go test ./... -v -race -count=1 -tags=integration

lint: ## Run go vet
	go vet ./...

fmt: ## Format the code with gofmt
	gofmt -s -w .

fmt-check: ## Fail if any file needs gofmt
	@files="$$(gofmt -l .)"; if [ -n "$$files" ]; then echo "Files needing gofmt:"; echo "$$files"; exit 1; fi

# Everything CI runs (.github/workflows/wip-ci.yml), plus the coverage floors:
# the full gate a lane must pass before it reports.
ci: fmt-check lint test cover-check build ## The full gate: fmt-check, vet, race tests, coverage floors, build

clean: ## Remove bin/
	rm -rf bin/

cover: ## Open the HTML coverage report
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out

cover-check: ## Check the per-package coverage floors
	@./scripts/check-coverage.sh
