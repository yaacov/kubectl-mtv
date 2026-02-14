# ============================================================================
# MCP E2E Test Suite – Makefile
#
# Environment:
#   Copy .env.example to .env and fill in credentials, or export the
#   variables in your shell before running.
#
# Server Lifecycle (3 steps):
#   1. Start server:    make server-start (binary) or server-start-image (container)
#   2. Run tests:       make test
#   3. Stop server:     make server-stop
#
# Quick workflow:
#   make test-full              # Start server, run all tests, stop server (binary)
#   make test-full-image        # Start server, run all tests, stop server (container)
#
# Run `make help` (or just `make`) to see all available targets.
# ============================================================================

SHELL   := /bin/bash
PYTHON  ?= python3
VENV    := .venv
PIP     := $(VENV)/bin/pip
PYTEST  := $(VENV)/bin/pytest

PYTEST_ARGS ?= -v -s --tb=short

# Scripts directory
SCRIPTS_DIR := scripts

# Default target
.DEFAULT_GOAL := help

# ── Setup ──────────────────────────────────────────────────────────────────

$(VENV)/bin/activate:
	$(PYTHON) -m venv $(VENV)

## install: Create virtualenv and install all dependencies (inc. dotenv)
.PHONY: install
install: $(VENV)/bin/activate
	$(PIP) install --upgrade pip
	$(PIP) install "."

# ── Server Management ──────────────────────────────────────────────────────

## server-start: Start MCP server (binary mode) in background
.PHONY: server-start
server-start:
	@$(SCRIPTS_DIR)/server-start.sh

## server-start-image: Start MCP server (container mode) in background
.PHONY: server-start-image
server-start-image:
	@$(SCRIPTS_DIR)/server-start-image.sh

## server-stop: Stop MCP server (binary mode)
.PHONY: server-stop
server-stop:
	@$(SCRIPTS_DIR)/server-stop.sh

## server-stop-image: Stop MCP server (container mode)
.PHONY: server-stop-image
server-stop-image:
	@$(SCRIPTS_DIR)/server-stop-image.sh

## server-stop-all: Stop all MCP servers (both binary and container)
.PHONY: server-stop-all
server-stop-all:
	@$(SCRIPTS_DIR)/server-stop.sh || true
	@$(SCRIPTS_DIR)/server-stop-image.sh || true

## server-status: Check if MCP server is running
.PHONY: server-status
server-status:
	@$(SCRIPTS_DIR)/server-status.sh

# ── Full suite ─────────────────────────────────────────────────────────────

## test: Run the full ordered test suite (requires running server)
.PHONY: test
test: install
	$(PYTEST) $(PYTEST_ARGS)

## test-full: Start server (binary), run tests, stop server
.PHONY: test-full
test-full: install server-stop
	@trap 'make server-stop' EXIT; $(SCRIPTS_DIR)/server-start.sh && $(PYTEST) $(PYTEST_ARGS)

## test-full-image: Start server (container), run tests, stop server (set MCP_IMAGE)
.PHONY: test-full-image
test-full-image: install server-stop-image
	@trap 'make server-stop-image' EXIT; $(SCRIPTS_DIR)/server-start-image.sh && $(PYTEST) $(PYTEST_ARGS)

# ── Per-concern targets ────────────────────────────────────────────────────

## test-setup: Run only MCP server startup + namespace creation tests
.PHONY: test-setup
test-setup: install
	$(PYTEST) $(PYTEST_ARGS) setup/

## test-providers: Run only provider create + read tests
.PHONY: test-providers
test-providers: install
	$(PYTEST) $(PYTEST_ARGS) providers/

## test-hosts: Run only ESXi host create + read tests
.PHONY: test-hosts
test-hosts: install
	$(PYTEST) $(PYTEST_ARGS) hosts/

## test-plans: Run only migration plan create + read tests
.PHONY: test-plans
test-plans: install
	$(PYTEST) $(PYTEST_ARGS) plans/

## test-mappings: Run only network/storage mapping read tests
.PHONY: test-mappings
test-mappings: install
	$(PYTEST) $(PYTEST_ARGS) mappings/

## test-inventory: Run only vSphere inventory read tests
.PHONY: test-inventory
test-inventory: install
	$(PYTEST) $(PYTEST_ARGS) inventory/

## test-health: Run only MTV health check tests
.PHONY: test-health
test-health: install
	$(PYTEST) $(PYTEST_ARGS) health/

## test-auth: Run only bearer-token authentication tests
.PHONY: test-auth
test-auth: install
	$(PYTEST) $(PYTEST_ARGS) auth/

# ── Cross-cutting ──────────────────────────────────────────────────────────

## test-create: Run all write (create) tests across every concern
.PHONY: test-create
test-create: install
	$(PYTEST) $(PYTEST_ARGS) -k "test_create or test_wait or test_print_banner or test_mcp_server_running or test_clean_namespace"

## test-read: Run all read tests across every concern
.PHONY: test-read
test-read: install
	$(PYTEST) $(PYTEST_ARGS) -k "not (test_create or test_wait or test_print_banner or test_mcp_server_running or test_clean_namespace or test_create_namespace)"

# ── Cleanup ────────────────────────────────────────────────────────────────

## clean: Remove virtualenv, server artifacts, and all __pycache__ / .pytest_cache dirs
.PHONY: clean
clean:
	@$(SCRIPTS_DIR)/server-stop.sh || true
	@$(SCRIPTS_DIR)/server-stop-image.sh || true
	rm -rf $(VENV) .pytest_cache __pycache__ .server.log .server.pid
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name .pytest_cache -exec rm -rf {} + 2>/dev/null || true

# ── Dynamic help ───────────────────────────────────────────────────────────

## help: Show this help message
.PHONY: help
help:
	@echo ""
	@echo "MCP E2E Test Suite"
	@echo "=================="
	@echo ""
	@echo "Server lifecycle is managed by scripts in scripts/"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | \
		sed 's/^## //' | \
		awk -F': ' '{printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}'
	@echo ""
