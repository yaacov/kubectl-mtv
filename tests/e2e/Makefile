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
	@echo "  test-version   - Run version command tests"
	@echo "  test-create    - Run all create tests"
	@echo "  test-get       - Run all get tests"
	@echo "  test-patch     - Run all patch tests"
	@echo ""
	@echo "Core Test Types:"
	@echo "  test-create-providers - Run provider creation tests"
	@echo "  test-create-hosts     - Run host creation tests"
	@echo "  test-create-plans     - Run plan creation tests"
	@echo "  test-create-mappings  - Run mapping creation tests"
	@echo "  test-create-hooks     - Run hook creation tests"
	@echo "  test-patch-providers  - Run provider patch tests"
	@echo "  test-patch-mappings   - Run mapping patch tests"
	@echo "  test-patch-plans      - Run plan patch tests"
	@echo "  test-get-inventory    - Run inventory retrieval tests"
	@echo ""
	@echo "Provider-Specific Tests:"
	@echo "  test-openshift - Run OpenShift provider tests"
	@echo "  test-vsphere   - Run VMware vSphere provider tests"
	@echo "  test-esxi      - Run VMware ESXi provider tests"
	@echo "  test-ovirt     - Run oVirt provider tests"
	@echo "  test-openstack - Run OpenStack provider tests"
	@echo "  test-ova       - Run OVA provider tests"
	@echo ""
	@echo "Utility targets:"
	@echo "  create-token   - Create service account and token for OpenShift tests"
	@echo "  clean-token    - Clean up test service account and token"
	@echo "  clean-test-namespaces - Clean up test namespaces"
	@echo "  list-ns        - List current test namespaces"
	@echo "  clean          - Clean up test artifacts"
	@echo "  clean-all      - Clean up everything including venv"
	@echo "  lint           - Run Python linting"
	@echo "  format         - Format Python code with black"
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

# Run all create tests
.PHONY: test-create
test-create: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) create/ -v -m create

# Run all get tests
.PHONY: test-get
test-get: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) get/ -v -m get

# Run all patch tests
.PHONY: test-patch
test-patch: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) patch/ -v -m patch

# Run provider creation tests
.PHONY: test-create-providers
test-create-providers: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) -v -m "create and providers"

# Run host creation tests
.PHONY: test-create-hosts
test-create-hosts: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) create/host/ -v -m "create and host"

# Run plan creation tests
.PHONY: test-create-plans
test-create-plans: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) create/plan/ -v -m "create and plan"

# Run mapping creation tests
.PHONY: test-create-mappings
test-create-mappings: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) create/mapping/ -v -m "create and mapping"

# Run hook creation tests
.PHONY: test-create-hooks
test-create-hooks: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) create/hook/ -v -m "create and hook"

# Run provider patch tests
.PHONY: test-patch-providers
test-patch-providers: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) -v -m "patch and providers"

# Run mapping patch tests
.PHONY: test-patch-mappings
test-patch-mappings: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) patch/mapping/ -v -m "patch and mapping"

# Run plan patch tests
.PHONY: test-patch-plans
test-patch-plans: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) patch/plan/ -v -m "patch and plan"

# Run inventory retrieval tests
.PHONY: test-get-inventory
test-get-inventory: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) get/inventory/ -v -m "get and inventory"

# Run OpenShift provider tests
.PHONY: test-openshift
test-openshift: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) -v -m openshift

# Run VMware vSphere provider tests
.PHONY: test-vsphere
test-vsphere: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) -v -m vsphere

# Run VMware ESXi provider tests
.PHONY: test-esxi
test-esxi: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) -v -m esxi

# Run oVirt provider tests
.PHONY: test-ovirt
test-ovirt: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) -v -m ovirt

# Run OpenStack provider tests
.PHONY: test-openstack
test-openstack: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) -v -m openstack

# Run OVA provider tests
.PHONY: test-ova
test-ova: setup check-binary
	@mkdir -p $(REPORTS_DIR)
	$(PYTEST) -v -m ova

# Cleanup test namespaces
.PHONY: clean-test-namespaces
clean-test-namespaces:
	@echo "Cleaning up test namespaces..."
	@kubectl get namespaces -o name | grep "namespace/kubectl-mtv-" | xargs -r kubectl delete --ignore-not-found=true
	@echo "Test namespaces cleaned up."

# List current test namespaces
.PHONY: list-ns
list-ns:
	@echo "Current test namespaces:"
	@kubectl get namespaces -o name | grep "kubectl-mtv-" | sed 's/namespace\///' || echo "No test namespaces found."

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

.PHONY: create-token
create-token:
	@echo "Creating service account and token for OpenShift tests..."
	@./create-test-service-account.sh
	@echo "Service account and token created."

.PHONY: clean-token
clean-token:
	@echo "Cleaning up test service account and token..."
	@kubectl delete serviceaccount test-mtv-sa --ignore-not-found=true
	@kubectl delete secret test-mtv-sa-token --ignore-not-found=true
	@echo "Test service account and token cleaned up."

.PHONY: lint
lint:
	@echo "Running Python linting..."
	@$(VENV_DIR)/bin/pip install --quiet flake8
	@$(VENV_DIR)/bin/flake8 . --exclude=$(VENV_DIR) --max-line-length=220
	@echo "Linting complete."

.PHONY: format
format:
	@echo "Formatting Python code with black..."
	@$(VENV_DIR)/bin/pip install --quiet black
	@$(VENV_DIR)/bin/black . --exclude $(VENV_DIR)
	@echo "Formatting complete."
