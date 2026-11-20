.PHONY: build run test test-fast test-integration lint fmt fmt-check ci clean cover cover-check

# The race detector needs cgo; some machines default it off.
RACE_ENV := CGO_ENABLED=1

build:
	go build -o bin/lazybrew ./cmd/lazybrew

# Build, then launch the TUI with your config.
run: build
	./bin/lazybrew

# All unit tests with the race detector (the release/lane gate's test step).
test:
	$(RACE_ENV) go test ./... -race -count=1

# Quick per-commit check: no race detector. Pass PKG to narrow it,
# e.g. `make test-fast PKG=./internal/gui/...`.
PKG ?= ./...
test-fast:
	go test $(PKG) -count=1

test-integration:
	$(RACE_ENV) go test ./... -v -race -count=1 -tags=integration

lint:
	go vet ./...

fmt:
	gofmt -s -w .

fmt-check:
	@files="$$(gofmt -l .)"; if [ -n "$$files" ]; then echo "Files needing gofmt:"; echo "$$files"; exit 1; fi

# Everything CI runs (.github/workflows/wip-ci.yml), plus the coverage floors:
# the full gate a lane must pass before it reports.
ci: fmt-check lint test cover-check build

clean:
	rm -rf bin/

cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out

cover-check:
	@./scripts/check-coverage.sh
