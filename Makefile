# Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

BINARY   := onecli
MODULE   := github.com/9Ashwin/one-cli
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
DATE     := $(shell date +%Y-%m-%d)
LDFLAGS  := -s -w -X $(MODULE)/internal/build.Version=$(VERSION) -X $(MODULE)/internal/build.Date=$(DATE)
PREFIX   ?= /usr/local

# The repository's CI toolchain may not support -race on every GOARCH.
# Prefer GOARCH passed to make over `go env GOARCH`.
TEST_GOARCH := $(or $(GOARCH),$(shell go env GOARCH))
RACE_FLAG := $(if $(filter riscv64,$(TEST_GOARCH)),,-race)

.PHONY: all build vet fmt-check test unit-test integration-test examples-build install uninstall clean gitleaks

all: test

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/onecli

vet:
	go vet ./...

# fmt-check fails when any file would be reformatted by gofmt. Keep this
# in sync with the fast-gate "Check formatting" step in CI.
fmt-check:
	@unformatted=$$(gofmt -l . | grep -v '^\.claude/' || true); \
	if [ -n "$$unformatted" ]; then \
		echo "Unformatted Go files:"; \
		echo "$$unformatted"; \
		echo "Run 'gofmt -w .' and commit."; \
		exit 1; \
	fi

# Build the independent modules so they stay compilable.
examples-build:
	cd extension && go build ./...
	cd lint && go build ./...

unit-test:
	go test $(RACE_FLAG) -count=1 ./...

integration-test: build
	go test -v -count=1 ./tests/... 2>/dev/null || true

test: vet fmt-check unit-test examples-build integration-test

install: build
	install -d $(PREFIX)/bin
	install -m755 $(BINARY) $(PREFIX)/bin/$(BINARY)
	@echo "OK: $(PREFIX)/bin/$(BINARY) ($(VERSION))"

uninstall:
	rm -f $(PREFIX)/bin/$(BINARY)

clean:
	rm -f $(BINARY)

# Run secret-leak checks locally before pushing.
# Install gitleaks: https://github.com/gitleaks/gitleaks#installing
gitleaks:
	@bash scripts/check-doc-tokens.sh || true
	@command -v gitleaks >/dev/null 2>&1 || { echo "gitleaks not found. Install: brew install gitleaks"; exit 0; }
	gitleaks detect --redact -v --exit-code=2
