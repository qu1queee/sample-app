#!/usr/bin/env bash
# build-buildah.sh — Reproduce Dockerfile.multistage using Buildah shell commands.
# No Dockerfile. No Docker daemon.
#
# Usage:
#   VERSION=1.0.0 ./build-buildah.sh
#
# On macOS: run this script inside a Buildah container (see workshop troubleshooting).
# On Linux: run directly, provided buildah is installed.
set -euo pipefail

VERSION="${VERSION:-dev}"
IMAGE_NAME="cc-week07-server"
IMAGE_TAG="${IMAGE_NAME}:${VERSION}"

# ---------------------------------------------------------------------------
# Stage 1: compile the Go binary
# ---------------------------------------------------------------------------
echo "==> [1/2] Building Go binary"

# buildah from pulls the base image and creates a mutable working container.
# The container name returned here (e.g. "golang-working-container") is used
# for all subsequent buildah commands — the same role that AS builder plays
# in a Dockerfile.
BUILDER=$(buildah from golang:1.23-bookworm)
echo "    builder container: ${BUILDER}"

# Equivalent to WORKDIR /app in a Dockerfile.
buildah config --workingdir /app "${BUILDER}"

# Copy source files into the working container.
# Equivalent to: COPY go.mod ./ then COPY . .
buildah copy "${BUILDER}" go.mod ./
buildah run "${BUILDER}" -- go mod download
buildah copy "${BUILDER}" main.go .

# Compile a fully static binary.
# CGO_ENABLED=0 removes the C runtime dependency — mandatory for distroless/static.
# GOOS=linux ensures the binary targets Linux even if you run Buildah on macOS.
# -w strips DWARF debugging info; -s strips the symbol table (~30% size reduction).
# Equivalent to the RUN go build step in Dockerfile.multistage.
buildah run \
  --env CGO_ENABLED=0 \
  --env GOOS=linux \
  "${BUILDER}" -- \
  go build -ldflags="-X main.version=${VERSION} -w -s" -o /app/server .

# ---------------------------------------------------------------------------
# Stage 2: assemble the minimal runtime image
# ---------------------------------------------------------------------------
echo "==> [2/2] Assembling runtime image"

# Create a second working container from the distroless static base image.
# distroless/static-debian12 contains only the C runtime and CA certificates —
# no shell, no package manager, no OS utilities. Equivalent to:
#   FROM gcr.io/distroless/static-debian12
RUNTIME=$(buildah from gcr.io/distroless/static-debian12)
echo "    runtime container: ${RUNTIME}"

# Copy ONLY the compiled binary from the builder container into the runtime container.
# --from copies between working containers rather than from the host filesystem.
# Equivalent to: COPY --from=builder /app/server /server
buildah copy --from "${BUILDER}" "${RUNTIME}" /app/server /server

# Run as the distroless nonroot user (UID 65532). Equivalent to: USER 65532:65532
buildah config --user "65532:65532" "${RUNTIME}"

# Declare the port the application listens on (OCI metadata, like EXPOSE).
buildah config --port 8080 "${RUNTIME}"

# Set the entrypoint. Equivalent to: ENTRYPOINT ["/server"]
buildah config --entrypoint '["/server"]' "${RUNTIME}"

# Add OCI-standard image labels for traceability.
buildah config \
  --label "org.opencontainers.image.version=${VERSION}" \
  --label "org.opencontainers.image.title=cc-week07-server" \
  "${RUNTIME}"

# ---------------------------------------------------------------------------
# Commit and clean up
# ---------------------------------------------------------------------------
echo "==> Committing final image as ${IMAGE_TAG}"
# Commit the runtime working container as a named OCI image in the local
# Buildah store. Equivalent to the implicit commit at the end of docker build.
buildah commit "${RUNTIME}" "${IMAGE_TAG}"

echo "==> Cleaning up working containers"
# Remove both working containers. The committed image stays in the Buildah store.
buildah rm "${BUILDER}" "${RUNTIME}"

echo ""
echo "Image created: ${IMAGE_TAG}"
echo ""
echo "Verify:                buildah images | grep ${IMAGE_NAME}"
echo "Load into Docker:      buildah push ${IMAGE_TAG} docker-daemon:${IMAGE_TAG}"
echo "Run with Docker:       docker run --rm -p 8080:8080 ${IMAGE_TAG}"
