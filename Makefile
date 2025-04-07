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
#   - CGO enabled
#   - musl-gcc package installed for static binary compilation
#
# Run `make install-tools` to install required development tools

VERSION_GIT := $(shell git describe --tags 2>/dev/null || echo "0.0.0-dev")
VERSION ?= ${VERSION_GIT}

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
	go build -ldflags='-X github.com/yaacov/kubectl-mtv/pkg/cmd.clientVersion=${VERSION}' -o kubectl-mtv $(kubemtv_cmd)

kubectl-mtv-static: $(kubemtv_cmd) $(kubemtv_pkg)
	CGO_ENABLED=1 CC=musl-gcc go build \
		-tags netgo \
		-ldflags '-linkmode external -extldflags "-static" -X github.com/yaacov/kubectl-mtv/pkg/cmd.clientVersion=${VERSION}' \
		-o kubectl-mtv \
		$(kubemtv_cmd)

.PHONY: lint
lint:
	go vet ./pkg/... ./cmd/...
	$(shell go env GOPATH)/bin/golangci-lint run ./pkg/... ./cmd/...

.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: dist
dist: kubectl-mtv
	tar -zcvf kubectl-mtv.tar.gz LICENSE kubectl-mtv
	sha256sum kubectl-mtv.tar.gz > kubectl-mtv.tar.gz.sha256sum

.PHONY: clean
clean:
	rm -f kubectl-mtv
	rm -f kubectl-mtv.tar.gz
	rm -f kubectl-mtv.tar.gz.sha256sum

.PHONY: test
test:
	go test -v -cover ./pkg/... ./cmd/...
	go test -coverprofile=coverage.out ./pkg/... ./cmd/...
	go tool cover -func=coverage.out
	@rm coverage.out