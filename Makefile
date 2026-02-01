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

# Path to forklift repository for verify-defaults target
# Override with: make verify-defaults FORKLIFT_PATH=/path/to/forklift
FORKLIFT_PATH ?= ../../kubev2v/forklift

all: kubectl-mtv

.PHONY: install-tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully. Make sure $(shell go env GOPATH)/bin is in your PATH to use them directly."

kubemtv_cmd := main.go
kubemtv_pkg := $(wildcard ./pkg/**/*.go)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

kubectl-mtv: clean $(kubemtv_cmd) $(kubemtv_pkg)
	@echo "Building for ${GOOS}/${GOARCH}"
	CGO_ENABLED=0 go build -ldflags='-X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}' -o kubectl-mtv $(kubemtv_cmd)

.PHONY: lint
lint:
	go vet ./pkg/... ./cmd/...
	$(shell go env GOPATH)/bin/golangci-lint run ./pkg/... ./cmd/...

.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

# Cross-compilation targets
.PHONY: build-linux-amd64
build-linux-amd64: $(kubemtv_cmd) $(kubemtv_pkg)
	@echo "Building for linux/amd64"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}' \
		-o kubectl-mtv-linux-amd64 \
		$(kubemtv_cmd)

.PHONY: build-linux-arm64
build-linux-arm64: $(kubemtv_cmd) $(kubemtv_pkg)
	@echo "Building for linux/arm64"
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}' \
		-o kubectl-mtv-linux-arm64 \
		$(kubemtv_cmd)

.PHONY: build-darwin-amd64
build-darwin-amd64: $(kubemtv_cmd) $(kubemtv_pkg)
	@echo "Building for darwin/amd64"
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}' \
		-o kubectl-mtv-darwin-amd64 \
		$(kubemtv_cmd)

.PHONY: build-darwin-arm64
build-darwin-arm64: $(kubemtv_cmd) $(kubemtv_pkg)
	@echo "Building for darwin/arm64"
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}' \
		-o kubectl-mtv-darwin-arm64 \
		$(kubemtv_cmd)

.PHONY: build-windows-amd64
build-windows-amd64: $(kubemtv_cmd) $(kubemtv_pkg)
	@echo "Building for windows/amd64"
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}' \
		-o kubectl-mtv-windows-amd64.exe \
		$(kubemtv_cmd)

# Build all platforms
.PHONY: build-all
build-all: clean build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

# Create release archives for all platforms
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

.PHONY: dist
dist: kubectl-mtv
	tar -zcvf kubectl-mtv.tar.gz LICENSE kubectl-mtv
	sha256sum kubectl-mtv.tar.gz > kubectl-mtv.tar.gz.sha256sum

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

.PHONY: test
test:
	go test -v -cover ./pkg/... ./cmd/...
	go test -coverprofile=coverage.out ./pkg/... ./cmd/...
	go tool cover -func=coverage.out
	@rm coverage.out

.PHONY: verify-defaults
verify-defaults:
	@echo "Verifying settings defaults against forklift..."
	@./scripts/verify-defaults.sh $(FORKLIFT_PATH)

.PHONY: test-e2e
test-e2e: kubectl-mtv
	@echo "Running e2e tests..."
	cd tests/e2e && python -m pytest -v

.PHONY: test-e2e-provider
test-e2e-provider: kubectl-mtv
	@echo "Running provider e2e tests..."
	cd tests/e2e && python -m pytest -v -m provider

.PHONY: test-cleanup
test-cleanup:
	@echo "Cleaning up test namespaces..."
	@kubectl get namespaces -o name | grep "namespace/kubectl-mtv-shared-" | xargs -r kubectl delete --ignore-not-found=true
	@echo "Test namespaces cleaned up."

.PHONY: test-list-namespaces
test-list-namespaces:
	@echo "Current test namespaces:"
	@kubectl get namespaces -o name | grep "kubectl-mtv-shared-" | sed 's/namespace\///' || echo "No test namespaces found."