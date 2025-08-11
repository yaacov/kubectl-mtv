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
	$(PYTHON) setup.py sdist bdist_wheel

.PHONY: build-sdist
build-sdist: clean
	$(PYTHON) setup.py sdist

.PHONY: build-wheel
build-wheel: clean
	$(PYTHON) setup.py bdist_wheel

.PHONY: upload-test
upload-test: build
	twine upload --repository testpypi $(DIST_DIR)/*

.PHONY: upload
upload: build
	twine upload $(DIST_DIR)/*

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)/
	rm -rf $(DIST_DIR)/
	rm -rf *.egg-info/
	find . -type f -name "*.pyc" -delete
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true

.PHONY: clean-all
clean-all: clean
	rm -f *-config-generated.json

.PHONY: config
config:
	$(PYTHON) generate_config.py