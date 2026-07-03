.PHONY: build test run install snapshot npm-stage clean

BIN := kyrc
PKG := ./cmd/kyrc

# Local dev build with version metadata from git.
build:
	go build -ldflags "-s -w \
		-X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev) \
		-X main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo none) \
		-X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" \
		-o $(BIN) $(PKG)

test:
	go test ./...

# Run the app locally.
run: build
	./$(BIN)

install:
	go install $(PKG)

# Cross-platform snapshot build via GoReleaser (no publish).
snapshot:
	goreleaser release --snapshot --clean

# Stage npm platform packages from dist/ (run after `make snapshot`).
npm-stage:
	node npm/scripts/build-packages.js

clean:
	rm -rf $(BIN) dist npm/packages
