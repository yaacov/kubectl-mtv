# ============================================================================
# MCP E2E Test Suite – Makefile
#
# Environment:
#   Copy .env.example to .env and fill in credentials, or export the
#   variables in your shell before running.
#
# Run `make help` (or just `make`) to see all available targets.
# ============================================================================

SHELL   := /bin/bash
PYTHON  ?= python3
VENV    := .venv
PIP     := $(VENV)/bin/pip
PYTEST  := $(VENV)/bin/pytest

PYTEST_ARGS ?= -v -s --tb=short

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

# ── Full suite ─────────────────────────────────────────────────────────────

## test: Run the full ordered test suite
.PHONY: test
test: install
	$(PYTEST) $(PYTEST_ARGS)

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

# ── Container image mode ──────────────────────────────────────────────
# Set MCP_IMAGE to run the tests against a container instead of the local binary.
# Example: make test-image MCP_IMAGE=quay.io/kubev2v/kubectl-mtv-mcp-server:latest

MCP_IMAGE ?=

## test-image: Run the full ordered test suite against a container image
.PHONY: test-image
test-image: install
	MCP_IMAGE=$(MCP_IMAGE) $(PYTEST) $(PYTEST_ARGS)

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

## clean: Remove virtualenv and all __pycache__ / .pytest_cache dirs
.PHONY: clean
clean:
	rm -rf $(VENV) .pytest_cache __pycache__
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
	@grep -E '^## ' $(MAKEFILE_LIST) | \
		sed 's/^## //' | \
		awk -F': ' '{printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
