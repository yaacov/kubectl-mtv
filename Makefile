# Copyright 2025 Yaacov Zamir <kobi.zamir@gmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Prerequisites:
#   - go 1.23 or higher
#   - curl or wget
#
# Run `make install-tools` to install required development tools

VERSION_GIT := $(shell git describe --tags 2>/dev/null || echo "0.0.0-dev")
VERSION ?= ${VERSION_GIT}

# Container image settings
IMAGE_REGISTRY ?= quay.io
IMAGE_ORG ?= kubev2v
IMAGE_NAME ?= kubectl-mtv-mcp-server
IMAGE_TAG ?= $(VERSION)
IMAGE ?= $(IMAGE_REGISTRY)/$(IMAGE_ORG)/$(IMAGE_NAME)
CONTAINER_ENGINE ?= $(shell command -v docker 2>/dev/null || command -v podman 2>/dev/null)

# Path to forklift repository for verify-defaults target
# Override with: make verify-defaults FORKLIFT_PATH=/path/to/forklift
FORKLIFT_PATH ?= ../../kubev2v/forklift

all: kubectl-mtv

## help: Show this help message
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## [a-zA-Z0-9_-]+:' $(MAKEFILE_LIST) | sort | \
		awk -F ': ' '{printf "  %-25s %s\n", substr($$1, 4), $$2}'

## install-tools: Install golangci-lint and other development tools
.PHONY: install-tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully. Make sure $(shell go env GOPATH)/bin is in your PATH to use them directly."

kubemtv_cmd := main.go
kubemtv_pkg := $(wildcard ./pkg/**/*.go)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

## kubectl-mtv: Build the kubectl-mtv binary for current platform
kubectl-mtv: clean $(kubemtv_cmd) $(kubemtv_pkg)
	@echo "Building for ${GOOS}/${GOARCH}"
	CGO_ENABLED=0 go build -ldflags='-X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}' -o kubectl-mtv $(kubemtv_cmd)

## lint: Run go vet and golangci-lint on codebase
.PHONY: lint
lint:
	go vet ./pkg/... ./cmd/...
	$(shell go env GOPATH)/bin/golangci-lint run ./pkg/... ./cmd/...

## fmt: Format Go code with go fmt
.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

## build-linux-amd64: Cross-compile for linux/amd64
.PHONY: build-linux-amd64
build-linux-amd64: $(kubemtv_cmd) $(kubemtv_pkg)
	@echo "Building for linux/amd64"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}' \
		-o kubectl-mtv-linux-amd64 \
		$(kubemtv_cmd)

## build-linux-arm64: Cross-compile for linux/arm64
.PHONY: build-linux-arm64
build-linux-arm64: $(kubemtv_cmd) $(kubemtv_pkg)
	@echo "Building for linux/arm64"
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}' \
		-o kubectl-mtv-linux-arm64 \
		$(kubemtv_cmd)

## build-darwin-amd64: Cross-compile for darwin/amd64
.PHONY: build-darwin-amd64
build-darwin-amd64: $(kubemtv_cmd) $(kubemtv_pkg)
	@echo "Building for darwin/amd64"
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}' \
		-o kubectl-mtv-darwin-amd64 \
		$(kubemtv_cmd)

## build-darwin-arm64: Cross-compile for darwin/arm64
.PHONY: build-darwin-arm64
build-darwin-arm64: $(kubemtv_cmd) $(kubemtv_pkg)
	@echo "Building for darwin/arm64"
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}' \
		-o kubectl-mtv-darwin-arm64 \
		$(kubemtv_cmd)

## build-windows-amd64: Cross-compile for windows/amd64
.PHONY: build-windows-amd64
build-windows-amd64: $(kubemtv_cmd) $(kubemtv_pkg)
	@echo "Building for windows/amd64"
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}' \
		-o kubectl-mtv-windows-amd64.exe \
		$(kubemtv_cmd)

## build-all: Build for all platforms (linux, darwin, windows)
.PHONY: build-all
build-all: clean build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

## dist-all: Create release archives and checksums for all platforms
.PHONY: dist-all
dist-all: build-all
	@echo "Creating release archives..."
	tar -zcvf kubectl-mtv-${VERSION}-linux-amd64.tar.gz LICENSE kubectl-mtv-linux-amd64
	tar -zcvf kubectl-mtv-${VERSION}-linux-arm64.tar.gz LICENSE kubectl-mtv-linux-arm64
	tar -zcvf kubectl-mtv-${VERSION}-darwin-amd64.tar.gz LICENSE kubectl-mtv-darwin-amd64
	tar -zcvf kubectl-mtv-${VERSION}-darwin-arm64.tar.gz LICENSE kubectl-mtv-darwin-arm64
	zip kubectl-mtv-${VERSION}-windows-amd64.zip LICENSE kubectl-mtv-windows-amd64.exe
	@echo "Generating individual checksums..."
	sha256sum kubectl-mtv-${VERSION}-linux-amd64.tar.gz > kubectl-mtv-${VERSION}-linux-amd64.tar.gz.sha256sum
	sha256sum kubectl-mtv-${VERSION}-linux-arm64.tar.gz > kubectl-mtv-${VERSION}-linux-arm64.tar.gz.sha256sum
	sha256sum kubectl-mtv-${VERSION}-darwin-amd64.tar.gz > kubectl-mtv-${VERSION}-darwin-amd64.tar.gz.sha256sum
	sha256sum kubectl-mtv-${VERSION}-darwin-arm64.tar.gz > kubectl-mtv-${VERSION}-darwin-arm64.tar.gz.sha256sum
	sha256sum kubectl-mtv-${VERSION}-windows-amd64.zip > kubectl-mtv-${VERSION}-windows-amd64.zip.sha256sum

## dist: Create tarball and checksum for current platform
.PHONY: dist
dist: kubectl-mtv
	tar -zcvf kubectl-mtv.tar.gz LICENSE kubectl-mtv
	sha256sum kubectl-mtv.tar.gz > kubectl-mtv.tar.gz.sha256sum

## clean: Remove build artifacts
.PHONY: clean
clean:
	rm -f kubectl-mtv
	rm -f kubectl-mtv-linux-amd64
	rm -f kubectl-mtv-linux-arm64
	rm -f kubectl-mtv-darwin-amd64
	rm -f kubectl-mtv-darwin-arm64
	rm -f kubectl-mtv-windows-amd64.exe
	rm -f kubectl-mtv.tar.gz
	rm -f kubectl-mtv.tar.gz.sha256sum
	rm -f kubectl-mtv-*-*.tar.gz
	rm -f kubectl-mtv-*-*.zip
	rm -f kubectl-mtv-*-*.tar.gz.sha256sum
	rm -f kubectl-mtv-*-*.zip.sha256sum

## test: Run unit tests with coverage
.PHONY: test
test:
	go test -v -coverprofile=coverage.out ./pkg/... ./cmd/...
	go tool cover -func=coverage.out
	@rm coverage.out

## dump-mcp-tools: Dump MCP tools/list response verbatim from kubectl-mtv MCP server
.PHONY: dump-mcp-tools
dump-mcp-tools: kubectl-mtv
	@echo "Dumping MCP tools/list response (stdio mode)..."
	@python3 scripts/dump-mcp-tools.py --stdio ./kubectl-mtv mcp-server

## verify-defaults: Verify settings defaults against forklift (FORKLIFT_PATH)
.PHONY: verify-defaults
verify-defaults:
	@echo "Verifying settings defaults against forklift..."
	@./scripts/verify-defaults.sh $(FORKLIFT_PATH)

## test-e2e-mcp: Run MCP e2e tests against the latest local build
.PHONY: test-e2e-mcp
test-e2e-mcp: kubectl-mtv
	@echo "Running MCP e2e tests against local build..."
	cd e2e/mcp && $(MAKE) test

## test-e2e-mcp-image: Run MCP e2e tests against a container image (MCP_IMAGE required)
.PHONY: test-e2e-mcp-image
test-e2e-mcp-image:
ifndef MCP_IMAGE
	$(error MCP_IMAGE is required. Example: make test-e2e-mcp-image MCP_IMAGE=$(IMAGE):$(IMAGE_TAG)-amd64)
endif
	@echo "Running MCP e2e tests against container image $(MCP_IMAGE)..."
	cd e2e/mcp && MCP_IMAGE=$(MCP_IMAGE) $(MAKE) test

## test-cleanup: Clean up test namespaces
.PHONY: test-cleanup
test-cleanup:
	@echo "Cleaning up test namespaces..."
	@kubectl get namespaces -o name | grep "namespace/kubectl-mtv-shared-" | xargs -r kubectl delete --ignore-not-found=true
	@echo "Test namespaces cleaned up."

## test-list-namespaces: List current test namespaces
.PHONY: test-list-namespaces
test-list-namespaces:
	@echo "Current test namespaces:"
	@kubectl get namespaces -o name | grep "kubectl-mtv-shared-" | sed 's/namespace\///' || echo "No test namespaces found."

# ---- Container image targets ----

## image-build-amd64: Build container image for linux/amd64
.PHONY: image-build-amd64
image-build-amd64:
	$(CONTAINER_ENGINE) build \
		--platform linux/amd64 \
		--build-arg TARGETARCH=amd64 \
		--build-arg VERSION=$(VERSION) \
		-f Containerfile \
		-t $(IMAGE):$(IMAGE_TAG)-amd64 \
		.

## image-build-arm64: Build container image for linux/arm64
.PHONY: image-build-arm64
image-build-arm64:
	$(CONTAINER_ENGINE) build \
		--platform linux/arm64 \
		--build-arg TARGETARCH=arm64 \
		--build-arg VERSION=$(VERSION) \
		-f Containerfile \
		-t $(IMAGE):$(IMAGE_TAG)-arm64 \
		.

## image-push-amd64: Push amd64 container image
.PHONY: image-push-amd64
image-push-amd64:
	$(CONTAINER_ENGINE) push $(IMAGE):$(IMAGE_TAG)-amd64

## image-push-arm64: Push arm64 container image
.PHONY: image-push-arm64
image-push-arm64:
	$(CONTAINER_ENGINE) push $(IMAGE):$(IMAGE_TAG)-arm64

## image-manifest: Create and push multi-arch manifest list
.PHONY: image-manifest
image-manifest:
	$(CONTAINER_ENGINE) manifest create --amend $(IMAGE):$(IMAGE_TAG) \
		$(IMAGE):$(IMAGE_TAG)-amd64 \
		$(IMAGE):$(IMAGE_TAG)-arm64
	$(CONTAINER_ENGINE) manifest push $(IMAGE):$(IMAGE_TAG)

## image-build-all: Build container images for all architectures
.PHONY: image-build-all
image-build-all: image-build-amd64 image-build-arm64

## image-push-all: Push all arch images and create multi-arch manifest
.PHONY: image-push-all
image-push-all: image-push-amd64 image-push-arm64 image-manifest
