PYTHON ?= python3
SRC_DIR = kubev2v
DIST_DIR = dist
BUILD_DIR = build

.PHONY: install
install:
	$(PYTHON) -m pip install -r requirements.txt
	$(PYTHON) -m pip install .

.PHONY: install-dev
install-dev:
	$(PYTHON) -m pip install -r requirements.txt
	$(PYTHON) -m pip install -r requirements-dev.txt
	$(PYTHON) -m pip install -e .

.PHONY: run-read
run-read:
	$(PYTHON) $(SRC_DIR)/kubectl_mtv_server.py

.PHONY: run-write
run-write:
	$(PYTHON) $(SRC_DIR)/kubectl_mtv_write_server.py

.PHONY: lint
lint:
	flake8 --max-line-length=360 $(SRC_DIR) *.py

.PHONY: format
format:
	black $(SRC_DIR) *.py

.PHONY: format-check
format-check:
	black --check $(SRC_DIR) *.py

.PHONY: build
build: clean
	$(PYTHON) -m build

.PHONY: build-sdist
build-sdist: clean
	$(PYTHON) -m build --sdist

.PHONY: build-wheel
build-wheel: clean
	$(PYTHON) -m build --wheel

.PHONY: upload-test
upload-test: build
	twine upload --repository testpypi $(DIST_DIR)/*

.PHONY: upload
upload: build
	twine upload $(DIST_DIR)/*

# Pyinstaller executable builds
MCP_SERVERS := kubectl-mtv-mcp kubectl-mtv-write-mcp kubevirt-mcp

# Individual MCP server targets
.PHONY: kubectl-mtv-mcp
kubectl-mtv-mcp: $(SRC_DIR)/kubectl_mtv_server.py hook-fastmcp.py
	@echo "Building kubectl-mtv-mcp server..."
	pyinstaller --onefile --additional-hooks-dir=. --hidden-import=fastmcp --strip --name kubectl-mtv-mcp $(SRC_DIR)/kubectl_mtv_server.py

.PHONY: kubectl-mtv-write-mcp
kubectl-mtv-write-mcp: $(SRC_DIR)/kubectl_mtv_write_server.py hook-fastmcp.py
	@echo "Building kubectl-mtv-write-mcp server..."
	pyinstaller --onefile --additional-hooks-dir=. --hidden-import=fastmcp --strip --name kubectl-mtv-write-mcp $(SRC_DIR)/kubectl_mtv_write_server.py

.PHONY: kubevirt-mcp
kubevirt-mcp: $(SRC_DIR)/kubevirt_server.py hook-fastmcp.py
	@echo "Building kubevirt-mcp server..."
	pyinstaller --onefile --additional-hooks-dir=. --hidden-import=fastmcp --strip --name kubevirt-mcp $(SRC_DIR)/kubevirt_server.py

# Build all MCP servers
.PHONY: build-servers
build-servers: kubectl-mtv-mcp kubectl-mtv-write-mcp kubevirt-mcp

# Package all MCP servers with LICENSE
.PHONY: dist-servers
dist-servers: build-servers
	@echo "Packaging MCP servers..."
	@mkdir -p servers-release
	@cp ../LICENSE servers-release/
	@cp $(DIST_DIR)/kubectl-mtv-mcp servers-release/
	@cp $(DIST_DIR)/kubectl-mtv-write-mcp servers-release/
	@cp $(DIST_DIR)/kubevirt-mcp servers-release/
	tar -zcvf kubectl-mtv-mcp-servers.tar.gz -C servers-release .
	sha256sum kubectl-mtv-mcp-servers.tar.gz > kubectl-mtv-mcp-servers.tar.gz.sha256sum
	@rm -rf servers-release

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)/
	rm -rf $(DIST_DIR)/
	rm -rf *.egg-info/
	rm -f kubectl-mtv-mcp-servers.tar.gz
	rm -f kubectl-mtv-mcp-servers.tar.gz.sha256sum
	rm -rf servers-release
	rm -f *.spec
	find . -type f -name "*.pyc" -delete
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
