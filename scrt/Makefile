VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BINARY  := scrt
GOFLAGS := -trimpath
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"
MODULE  := github.com/alexrf45/scrt

.PHONY: all build test vet lint clean install docker-build docker-test help

all: vet test build

## build: Compile the binary with version injection (CI-2)
build:
	go build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY) ./cmd/scrt

## test: Run all tests with race detector (G-3)
test:
	go test -race -count=1 ./...

## vet: Run go vet on all packages (G-1)
vet:
	go vet ./...

## lint: Run golangci-lint (G-2)
lint:
	golangci-lint run ./...

## fmt: Format code with gofmt
fmt:
	gofmt -w .

## clean: Remove build artifacts
clean:
	rm -rf bin/

## install: Build and copy binary to ~/.local/bin
install: build
	mkdir -p $(HOME)/.local/bin
	cp bin/$(BINARY) $(HOME)/.local/bin/$(BINARY)

# ============================================================================
# Docker-based build targets (bypass local toolchain issues)
# ============================================================================

## docker-build: Build binary inside Docker, output to bin/
docker-build:
	@mkdir -p bin
	DOCKER_BUILDKIT=1 docker build \
		-f Dockerfile.build \
		--build-arg VERSION=$(VERSION) \
		--network=host \
		--target binary \
		--output type=local,dest=bin/ \
		.
	@echo "Binary: bin/scrt ($(VERSION))"

## docker-test: Run vet + tests inside Docker
docker-test:
	DOCKER_BUILDKIT=1 docker build \
		-f Dockerfile.build \
		--build-arg VERSION=$(VERSION) \
		--network=host \
		--target tester \
		.

## docker-all: Full CI pipeline in Docker (test + build)
docker-all: docker-test docker-build

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
