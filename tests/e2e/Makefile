# Makefile for kubectl-mtv e2e tests

# Python virtual environment
VENV_DIR = venv
PYTHON = $(VENV_DIR)/bin/python
PIP = $(VENV_DIR)/bin/pip
PYTEST = $(VENV_DIR)/bin/pytest

# Test configuration
TEST_DIR = .
REPORTS_DIR = reports
HTML_REPORT = $(REPORTS_DIR)/report.html
JSON_REPORT = $(REPORTS_DIR)/report.json
COVERAGE_REPORT = $(REPORTS_DIR)/coverage.html

# Default target
.PHONY: help
help:
	@echo "kubectl-mtv E2E Test Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  setup          - Set up Python virtual environment and install dependencies"
	@echo "  test           - Run all tests"
	@echo ""
	@echo "Test Categories:"
	@echo "  test-version   - Run version command tests only"
	@echo "  test-providers - Run all provider tests"
	@echo "  test-inventory - Run inventory command tests only" 
	@echo "  test-plans     - Run migration plan tests only"
	@echo "  test-mappings  - Run mapping tests only"
	@echo ""
	@echo "Provider Tests:"
	@echo "  test-openshift - Run OpenShift provider tests only"
	@echo "  test-vsphere   - Run VMware vSphere provider tests only"
	@echo "  test-ovirt     - Run oVirt provider tests only"
	@echo "  test-openstack - Run OpenStack provider tests only"
	@echo "  test-ova       - Run OVA provider tests only"
	@echo "  test-errors    - Run provider error/edge case tests only"
	@echo ""
	@echo "Test Options:"
	@echo "  test-no-creds  - Run tests that don't require credentials"
	@echo "  test-fast      - Run tests in parallel"
	@echo "  test-report    - Run tests and generate HTML report"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo ""
	@echo "Utility targets:"
	@echo "  cleanup        - Clean up test namespaces"
	@echo "  list-ns        - List current test namespaces"
	@echo "  clean          - Clean up test artifacts"
	@echo "  clean-all      - Clean up everything including venv"
	@echo "  lint           - Run Python linting"
	@echo "  check-cluster  - Check cluster connectivity and permissions"
	@echo "  check-binary   - Check if kubectl-mtv binary is available"

# Setup virtual environment and dependencies
.PHONY: setup
setup: $(VENV_DIR)/bin/activate

$(VENV_DIR)/bin/activate: requirements.txt
	python3 -m venv $(VENV_DIR)
	$(PIP) install --upgrade pip
	$(PIP) install -r requirements.txt
	touch $(VENV_DIR)/bin/activate

# Run all tests
.PHONY: test
test: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) $(TEST_DIR) -v

# Run version tests only
.PHONY: test-version
test-version: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) version/ -v -m version

# Run inventory tests only
.PHONY: test-inventory
test-inventory: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) inventory/ -v -m inventory

# Run plan tests only
.PHONY: test-plans
test-plans: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) plans/ -v -m plan

# Run mapping tests only
.PHONY: test-mappings
test-mappings: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) mappings/ -v -m mapping

# Run provider tests only
.PHONY: test-providers
test-providers: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) providers/ -v -m provider

# Run OpenShift provider tests only
.PHONY: test-openshift
test-openshift: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) providers/ -v -m openshift

# Run VMware vSphere provider tests only
.PHONY: test-vsphere
test-vsphere: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) providers/ -v -m vsphere

# Run oVirt provider tests only
.PHONY: test-ovirt
test-ovirt: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) providers/ -v -m ovirt

# Run OpenStack provider tests only
.PHONY: test-openstack
test-openstack: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) providers/ -v -m openstack

# Run OVA provider tests only
.PHONY: test-ova
test-ova: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) providers/ -v -m ova

# Run provider error/edge case tests only
.PHONY: test-errors
test-errors: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) providers/ -v -m error_cases

# Run tests that don't require credentials
.PHONY: test-no-creds
test-no-creds: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) -v -m "not requires_credentials"

# Run tests in parallel
.PHONY: test-fast
test-fast: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) $(TEST_DIR) -v -n auto

# Cleanup test namespaces
.PHONY: cleanup
cleanup:
	@echo "Cleaning up test namespaces..."
	@kubectl get namespaces -o name | grep "namespace/kubectl-mtv-shared-" | xargs -r kubectl delete --ignore-not-found=true
	@echo "Test namespaces cleaned up."

# List current test namespaces
.PHONY: list-ns
list-ns:
	@echo "Current test namespaces:"
	@kubectl get namespaces -o name | grep "kubectl-mtv-shared-" | sed 's/namespace\///' || echo "No test namespaces found."

# Run tests and generate HTML report
.PHONY: test-report
test-report: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) $(TEST_DIR) -v \
		--html=$(HTML_REPORT) \
		--self-contained-html \
		--json-report \
		--json-report-file=$(JSON_REPORT)
	@echo "HTML report generated: $(HTML_REPORT)"
	@echo "JSON report generated: $(JSON_REPORT)"

# Run tests with coverage
.PHONY: test-coverage
test-coverage: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PIP) install pytest-cov
	$(PYTEST) $(TEST_DIR) -v \
		--cov=. \
		--cov-report=html:$(COVERAGE_REPORT) \
		--cov-report=term-missing
	@echo "Coverage report generated: $(COVERAGE_REPORT)"

# Check cluster connectivity and permissions
.PHONY: check-cluster
check-cluster:
	@echo "Checking cluster connectivity..."
	@kubectl cluster-info --request-timeout=10s > /dev/null || (echo "❌ Not connected to cluster" && exit 1)
	@echo "✅ Connected to cluster"
	@echo "Checking admin permissions..."
	@kubectl auth can-i '*' '*' --all-namespaces > /dev/null || (echo "❌ No admin permissions" && exit 1)
	@echo "✅ Admin permissions confirmed"
	@echo "Current context: $$(kubectl config current-context)"

# Check if kubectl-mtv binary is available
.PHONY: check-binary
check-binary:
	@echo "Checking for kubectl-mtv binary..."
	@which kubectl-mtv > /dev/null 2>&1 || \
		(test -f ../../kubectl-mtv && echo "✅ Found kubectl-mtv in project root") || \
		(test -f ../../../kubectl-mtv && echo "✅ Found kubectl-mtv in project root") || \
		(echo "❌ kubectl-mtv binary not found. Run 'make' in project root to build it." && exit 1)
	@echo "✅ kubectl-mtv binary is available"

# Run Python linting
.PHONY: lint
lint: setup
	$(PIP) install flake8 black isort
	$(VENV_DIR)/bin/black --check .
	$(VENV_DIR)/bin/isort --check-only .
	$(VENV_DIR)/bin/flake8 .

# Format Python code
.PHONY: format
format: setup
	$(PIP) install black isort
	$(VENV_DIR)/bin/black .
	$(VENV_DIR)/bin/isort .

# Clean test artifacts
.PHONY: clean
clean:
	rm -rf $(REPORTS_DIR)
	rm -rf .pytest_cache
	rm -rf __pycache__
	find . -name "*.pyc" -delete
	find . -name "*.pyo" -delete

# Clean everything including virtual environment
.PHONY: clean-all
clean-all: clean
	rm -rf $(VENV_DIR)

# Run pre-commit checks
.PHONY: pre-commit
pre-commit: check-cluster check-binary lint test

# Quick smoke test
.PHONY: smoke-test
smoke-test: setup check-binary
	@echo "Running smoke test..."
	$(PYTEST) version/test_version.py::TestVersion::test_version_command_basic -v

# Install development dependencies
.PHONY: dev-setup
dev-setup: setup
	$(PIP) install black isort flake8 pytest-cov

# Show test environment info
.PHONY: info
info:
	@echo "Test Environment Information"
	@echo "============================"
	@echo "Python: $$(python3 --version 2>&1)"
	@echo "Virtual env: $(VENV_DIR)"
	@echo "Test directory: $(TEST_DIR)"
	@echo "Reports directory: $(REPORTS_DIR)"
	@echo ""
	@echo "Test Configuration"
	@echo "=================="
	@echo "- All tests run in shared namespace (preserved for debugging)"
	@echo "- Resources are cleaned up automatically"
	@echo "- Use 'make cleanup' to remove test namespaces"
	@echo "- Use 'make list-ns' to see current test namespaces"
	@echo ""
	@echo "Test Files by Category"
	@echo "======================"
	@echo "Version Tests:"
	@echo "  version/test_version.py           - Version command tests"
	@echo ""
	@echo "Provider Tests:"
	@echo "  providers/test_openshift.py       - OpenShift provider tests"
	@echo "  providers/test_vsphere.py         - VMware vSphere provider tests"
	@echo "  providers/test_ovirt.py           - oVirt provider tests"
	@echo "  providers/test_openstack.py       - OpenStack provider tests"
	@echo "  providers/test_ova.py             - OVA provider tests"
	@echo "  providers/test_errors.py          - Error condition tests"
	@echo ""
	@echo "Inventory Tests:"
	@echo "  inventory/test_inventory.py       - Inventory listing and querying"
	@echo ""
	@echo "Plan Tests:"
	@echo "  plans/test_plans.py               - Migration plan operations"
	@echo ""
	@echo "Mapping Tests:"
	@echo "  mappings/test_mappings.py         - Network and storage mappings"
	@echo ""
	@echo "Cluster Information"
	@echo "==================="
	@kubectl cluster-info --request-timeout=5s 2>/dev/null || echo "Not connected to cluster"
	@echo ""
	@echo "kubectl-mtv binary"
	@echo "=================="
	@which kubectl-mtv 2>/dev/null || echo "Not in PATH"
	@test -f ../../kubectl-mtv && echo "Found in project root: ../../kubectl-mtv" || true
	@test -f ../../../kubectl-mtv && echo "Found in project root: ../../../kubectl-mtv" || true
