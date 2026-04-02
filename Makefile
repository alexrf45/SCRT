VERSION    ?= $(shell git -C $(CURDIR) describe --tags --always --dirty 2>/dev/null || echo dev)
PREFIX     := scrt
PLATFORM   := linux/amd64

BASE_IMAGE := $(PREFIX)/base:$(VERSION)
WEB_IMAGE  := $(PREFIX)/web:$(VERSION)
AD_IMAGE   := $(PREFIX)/ad:$(VERSION)
CTF_IMAGE  := $(PREFIX)/ctf:$(VERSION)

DOCKER     := DOCKER_BUILDKIT=1 docker
BUILD_OPTS := --platform $(PLATFORM) --build-arg VERSION=$(VERSION)

SCRT       := scrt/bin/scrt
TEST_DIR   := $(CURDIR)/tests

.DEFAULT_GOAL := help

.PHONY: build-scrt build-base build-web build-ad build-ctf build-all \
        smoke-web smoke-ad smoke-ctf test-all test-scrt \
        lint ci images clean-images clean help

# ============================================================================
# Help
# ============================================================================

## help:         Show this help
help:
	@printf '\nUsage: make <target>\n\n'
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
	@printf '\n'

# ============================================================================
# scrt binary
# ============================================================================

## build-scrt:   Compile the scrt binary → scrt/bin/scrt
build-scrt:
	@$(MAKE) -C scrt build

# Phony prerequisite so dependent targets always check the binary is built
$(SCRT):
	@$(MAKE) -C scrt build

# ============================================================================
# Image builds
# ============================================================================

## build-base:   Build the base Kali image          (Dockerfile)
build-base:
	@printf '==> [base] building %s\n' "$(BASE_IMAGE)"
	@$(DOCKER) build $(BUILD_OPTS) -t $(BASE_IMAGE) -f Dockerfile .

## build-web:    Build the web/bug-bounty image      (Dockerfile.web)
build-web:
	@printf '==> [web] building %s\n' "$(WEB_IMAGE)"
	@$(DOCKER) build $(BUILD_OPTS) -t $(WEB_IMAGE) -f Dockerfile.web .

## build-ad:     Build the AD/internal network image (Dockerfile.ad)
build-ad:
	@printf '==> [ad] building %s\n' "$(AD_IMAGE)"
	@$(DOCKER) build $(BUILD_OPTS) -t $(AD_IMAGE) -f Dockerfile.ad .

## build-ctf:    Build the CTF image                 (Dockerfile.ctf)
build-ctf:
	@printf '==> [ctf] building %s\n' "$(CTF_IMAGE)"
	@$(DOCKER) build $(BUILD_OPTS) -t $(CTF_IMAGE) -f Dockerfile.ctf .

## build-all:    Build web + ad + ctf images in parallel
build-all:
	@$(MAKE) -j3 build-web build-ad build-ctf

# ============================================================================
# Smoke tests
#
# Each target bind-mounts tests/ into the container and runs a POSIX sh
# script — no Makefile shell-escaping needed, easy to edit independently.
# Images are built first if the tags don't already exist.
# ============================================================================

## smoke-web:    Tool presence check — web image
smoke-web: build-web
	@printf '==> [web] smoke test\n'
	@docker run --rm \
		-v "$(TEST_DIR):/tests:ro" \
		$(WEB_IMAGE) sh /tests/smoke-web.sh

## smoke-ad:     Tool presence check — AD image
smoke-ad: build-ad
	@printf '==> [ad] smoke test\n'
	@docker run --rm \
		-v "$(TEST_DIR):/tests:ro" \
		$(AD_IMAGE) sh /tests/smoke-ad.sh

## smoke-ctf:    Tool presence check — CTF image
smoke-ctf: build-ctf
	@printf '==> [ctf] smoke test\n'
	@docker run --rm \
		-v "$(TEST_DIR):/tests:ro" \
		$(CTF_IMAGE) sh /tests/smoke-ctf.sh

## test-all:     Run smoke tests for all three scenario images
test-all: smoke-web smoke-ad smoke-ctf

# ============================================================================
# scrt integration — verify the binary starts and sees Docker
# ============================================================================

## test-scrt:    Verify scrt binary runs and can reach the Docker socket
test-scrt: $(SCRT)
	@printf '==> [scrt] integration check\n'
	@$(SCRT) version
	@printf '[PASS] scrt binary OK\n'

# ============================================================================
# Dockerfile linting (optional — no-op if hadolint is absent)
# ============================================================================

## lint:          Lint all Dockerfiles with hadolint
lint:
	@if command -v hadolint >/dev/null 2>&1; then \
		for f in Dockerfile Dockerfile.web Dockerfile.ad Dockerfile.ctf \
		         scrt/Dockerfile scrt/Dockerfile.build; do \
			printf '==> linting %s\n' "$$f"; \
			hadolint "$$f" || true; \
		done; \
	else \
		printf 'hadolint not found — skipping lint\n'; \
		printf 'install: https://github.com/hadolint/hadolint\n'; \
	fi

# ============================================================================
# Full local CI pipeline — run this before pushing to a feature branch
# ============================================================================

## ci:            build-scrt → build-all → test-all
ci: build-scrt build-all test-all
	@printf '\n==> All checks passed — safe to branch.\n'

# ============================================================================
# Utilities
# ============================================================================

## images:        List locally built scenario images
images:
	@docker images --filter "reference=$(PREFIX)/*" \
		--format 'table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedSince}}'

## clean-images:  Remove locally built scenario images
clean-images:
	@docker rmi -f \
		$(WEB_IMAGE) $(AD_IMAGE) $(CTF_IMAGE) $(BASE_IMAGE) \
		2>/dev/null || true
	@printf '==> images removed\n'

## clean:         Remove binary artifacts and locally built images
clean: clean-images
	@$(MAKE) -C scrt clean
