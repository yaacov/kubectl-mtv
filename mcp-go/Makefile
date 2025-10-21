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
#   - golangci-lint (run `make install-tools` to install)
#
# Quick start:
#   make build                      # Build all servers for current platform
#   make lint                       # Run linters
#   make fmt                        # Format code
#   make build-all                  # Build all servers for all platforms

VERSION_GIT := $(shell git describe --tags 2>/dev/null || echo "0.0.0-dev")
VERSION ?= ${VERSION_GIT}

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOPATH := $(shell go env GOPATH)

# Source files
PKG_FILES := $(shell find pkg -name '*.go')
READ_SERVER_FILES := $(shell find cmd/read-server -name '*.go')
WRITE_SERVER_FILES := $(shell find cmd/write-server -name '*.go')
VIRTCTL_SERVER_FILES := $(shell find cmd/virtctl-server -name '*.go')

# Build flags
LDFLAGS := -s -w -X main.Version=${VERSION}
BUILD_FLAGS := -trimpath

.PHONY: all
all: build

# Install development tools
.PHONY: install-tools
install-tools:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Build all servers for current platform
.PHONY: build
build: clean build-kubectl-mtv-mcp build-kubectl-mtv-write-mcp build-virtctl-mcp

# Individual server builds
.PHONY: build-kubectl-mtv-mcp
build-kubectl-mtv-mcp: $(PKG_FILES) $(READ_SERVER_FILES)
	@mkdir -p bin
	@CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o bin/kubectl-mtv-mcp ./cmd/read-server

.PHONY: build-kubectl-mtv-write-mcp
build-kubectl-mtv-write-mcp: $(PKG_FILES) $(WRITE_SERVER_FILES)
	@mkdir -p bin
	@CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o bin/kubectl-mtv-write-mcp ./cmd/write-server

.PHONY: build-virtctl-mcp
build-virtctl-mcp: $(PKG_FILES) $(VIRTCTL_SERVER_FILES)
	@mkdir -p bin
	@CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o bin/virtctl-mcp ./cmd/virtctl-server

# Code quality
.PHONY: lint
lint:
	@go vet ./pkg/... ./cmd/...
	@$(GOPATH)/bin/golangci-lint run ./pkg/... ./cmd/...

.PHONY: fmt
fmt:
	@go fmt ./pkg/... ./cmd/...

.PHONY: vet
vet:
	@go vet ./pkg/... ./cmd/...

# Testing
.PHONY: test
test:
	@go test -v -cover ./pkg/... ./cmd/...

.PHONY: test-coverage
test-coverage:
	@go test -coverprofile=coverage.out ./pkg/... ./cmd/...
	@go tool cover -func=coverage.out

# Multi-architecture builds
.PHONY: build-linux-amd64
build-linux-amd64: $(PKG_FILES) $(READ_SERVER_FILES) $(WRITE_SERVER_FILES) $(VIRTCTL_SERVER_FILES)
	@mkdir -p dist/linux-amd64
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/linux-amd64/kubectl-mtv-mcp ./cmd/read-server
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/linux-amd64/kubectl-mtv-write-mcp ./cmd/write-server
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/linux-amd64/virtctl-mcp ./cmd/virtctl-server

.PHONY: build-linux-arm64
build-linux-arm64: $(PKG_FILES) $(READ_SERVER_FILES) $(WRITE_SERVER_FILES) $(VIRTCTL_SERVER_FILES)
	@mkdir -p dist/linux-arm64
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/linux-arm64/kubectl-mtv-mcp ./cmd/read-server
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/linux-arm64/kubectl-mtv-write-mcp ./cmd/write-server
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/linux-arm64/virtctl-mcp ./cmd/virtctl-server

.PHONY: build-darwin-amd64
build-darwin-amd64: $(PKG_FILES) $(READ_SERVER_FILES) $(WRITE_SERVER_FILES) $(VIRTCTL_SERVER_FILES)
	@mkdir -p dist/darwin-amd64
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/darwin-amd64/kubectl-mtv-mcp ./cmd/read-server
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/darwin-amd64/kubectl-mtv-write-mcp ./cmd/write-server
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/darwin-amd64/virtctl-mcp ./cmd/virtctl-server

.PHONY: build-darwin-arm64
build-darwin-arm64: $(PKG_FILES) $(READ_SERVER_FILES) $(WRITE_SERVER_FILES) $(VIRTCTL_SERVER_FILES)
	@mkdir -p dist/darwin-arm64
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/darwin-arm64/kubectl-mtv-mcp ./cmd/read-server
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/darwin-arm64/kubectl-mtv-write-mcp ./cmd/write-server
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/darwin-arm64/virtctl-mcp ./cmd/virtctl-server

.PHONY: build-windows-amd64
build-windows-amd64: $(PKG_FILES) $(READ_SERVER_FILES) $(WRITE_SERVER_FILES) $(VIRTCTL_SERVER_FILES)
	@mkdir -p dist/windows-amd64
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/windows-amd64/kubectl-mtv-mcp.exe ./cmd/read-server
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/windows-amd64/kubectl-mtv-write-mcp.exe ./cmd/write-server
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a $(BUILD_FLAGS) -ldflags='$(LDFLAGS)' -o dist/windows-amd64/virtctl-mcp.exe ./cmd/virtctl-server

.PHONY: build-all
build-all: clean build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

# Create release archives
.PHONY: dist-all
dist-all: build-all
	@cd dist/linux-amd64 && tar -czf ../mcp-servers-${VERSION}-linux-amd64.tar.gz * && cd ../..
	@cd dist/linux-arm64 && tar -czf ../mcp-servers-${VERSION}-linux-arm64.tar.gz * && cd ../..
	@cd dist/darwin-amd64 && tar -czf ../mcp-servers-${VERSION}-darwin-amd64.tar.gz * && cd ../..
	@cd dist/darwin-arm64 && tar -czf ../mcp-servers-${VERSION}-darwin-arm64.tar.gz * && cd ../..
	@cd dist/windows-amd64 && zip -q -r ../mcp-servers-${VERSION}-windows-amd64.zip * && cd ../..
	@cd dist && sha256sum mcp-servers-${VERSION}-linux-amd64.tar.gz > mcp-servers-${VERSION}-linux-amd64.tar.gz.sha256sum
	@cd dist && sha256sum mcp-servers-${VERSION}-linux-arm64.tar.gz > mcp-servers-${VERSION}-linux-arm64.tar.gz.sha256sum
	@cd dist && sha256sum mcp-servers-${VERSION}-darwin-amd64.tar.gz > mcp-servers-${VERSION}-darwin-amd64.tar.gz.sha256sum
	@cd dist && sha256sum mcp-servers-${VERSION}-darwin-arm64.tar.gz > mcp-servers-${VERSION}-darwin-arm64.tar.gz.sha256sum
	@cd dist && sha256sum mcp-servers-${VERSION}-windows-amd64.zip > mcp-servers-${VERSION}-windows-amd64.zip.sha256sum

.PHONY: dist
dist: build
	@mkdir -p dist
	@tar -czf dist/mcp-servers-${VERSION}-${GOOS}-${GOARCH}.tar.gz -C bin .
	@cd dist && sha256sum mcp-servers-${VERSION}-${GOOS}-${GOARCH}.tar.gz > mcp-servers-${VERSION}-${GOOS}-${GOARCH}.tar.gz.sha256sum && cd ..

# Installation
.PHONY: install
install: build
	@cp bin/kubectl-mtv-mcp $(GOPATH)/bin/
	@cp bin/kubectl-mtv-write-mcp $(GOPATH)/bin/
	@cp bin/virtctl-mcp $(GOPATH)/bin/

# Run servers
.PHONY: run-kubectl-mtv-mcp
run-kubectl-mtv-mcp: build-kubectl-mtv-mcp
	@./bin/kubectl-mtv-mcp

.PHONY: run-kubectl-mtv-write-mcp
run-kubectl-mtv-write-mcp: build-kubectl-mtv-write-mcp
	@./bin/kubectl-mtv-write-mcp

.PHONY: run-virtctl-mcp
run-virtctl-mcp: build-virtctl-mcp
	@./bin/virtctl-mcp

# Cleanup
.PHONY: clean
clean:
	@rm -rf bin/ dist/ coverage.out

.PHONY: clean-all
clean-all: clean
	@go clean -cache -testcache -modcache

# Development helpers
.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: verify
verify: fmt vet lint test

.PHONY: deps
deps:
	@go mod download
