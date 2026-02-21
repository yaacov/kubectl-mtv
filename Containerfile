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

# ---- Builder stage (runs on native platform, cross-compiles for target) ----
FROM --platform=$BUILDPLATFORM registry.access.redhat.com/ubi9/go-toolset:latest AS builder

ARG TARGETARCH=amd64
ARG VERSION=0.0.0-dev

USER root
WORKDIR /build

# Copy go module files first for better layer caching
COPY go.mod go.sum ./
COPY vendor/ vendor/

# Copy source code
COPY main.go ./
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build kubectl-mtv
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -a \
    -ldflags "-s -w -X github.com/yaacov/kubectl-mtv/cmd.clientVersion=${VERSION}" \
    -o kubectl-mtv \
    main.go

# ---- Runtime stage ----
FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

ARG TARGETARCH=amd64

# Copy binary from builder (set execute permissions during copy)
COPY --from=builder --chmod=755 /build/kubectl-mtv /usr/local/bin/kubectl-mtv

# --- Environment variables ---
# SSE server settings
ENV MCP_HOST="0.0.0.0" \
    MCP_PORT="8080" \
    MCP_OUTPUT_FORMAT="text" \
    MCP_MAX_RESPONSE_CHARS="0" \
    MCP_READ_ONLY="false"

# TLS settings (optional - provide paths to enable TLS)
ENV MCP_CERT_FILE="" \
    MCP_KEY_FILE=""

# Kubernetes authentication (optional - override via HTTP headers in SSE mode)
ENV MCP_KUBE_SERVER="" \
    MCP_KUBE_TOKEN="" \
    MCP_KUBE_INSECURE=""

USER 1001
WORKDIR /home/mtv

EXPOSE 8080

ENTRYPOINT ["/bin/sh", "-c", "\
  exec kubectl-mtv mcp-server --sse \
    --host \"${MCP_HOST}\" \
    --port \"${MCP_PORT}\" \
    --output-format \"${MCP_OUTPUT_FORMAT}\" \
    ${MCP_MAX_RESPONSE_CHARS:+--max-response-chars \"${MCP_MAX_RESPONSE_CHARS}\"} \
    ${MCP_CERT_FILE:+--cert-file \"${MCP_CERT_FILE}\"} \
    ${MCP_KEY_FILE:+--key-file \"${MCP_KEY_FILE}\"} \
    ${MCP_KUBE_SERVER:+--server \"${MCP_KUBE_SERVER}\"} \
    ${MCP_KUBE_TOKEN:+--token \"${MCP_KUBE_TOKEN}\"} \
    $([ \"${MCP_KUBE_INSECURE}\" = \"true\" ] && echo --insecure-skip-tls-verify) \
    $([ \"${MCP_READ_ONLY}\" = \"true\" ] && echo --read-only)"]

# Labels at the end for better readability
LABEL name="kubectl-mtv-mcp-server" \
      summary="kubectl-mtv MCP server for AI-assisted VM migrations" \
      description="MCP (Model Context Protocol) server exposing kubectl-mtv migration toolkit for AI assistants. Runs in SSE mode over HTTP." \
      io.k8s.display-name="kubectl-mtv MCP Server" \
      io.k8s.description="MCP server for kubectl-mtv providing AI-assisted VM migration capabilities via SSE transport." \
      maintainer="Yaacov Zamir <kobi.zamir@gmail.com>"
