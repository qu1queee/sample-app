# Stage 1: build — full Go toolchain, ~900 MB
FROM golang:1.23-bookworm AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-X main.version=${VERSION} -w -s" \
    -o server .
# Create an empty /data directory; when a PVC is mounted at /data it replaces this.
RUN mkdir /data

# Stage 2: runtime — distroless static, ~2 MB, no shell, no package manager
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/server /server
# Copy the empty /data directory with nonroot ownership so the process can write to it.
COPY --from=builder --chown=65532:65532 /data /data

# UID 65532 is the "nonroot" user defined in distroless images.
USER 65532:65532

EXPOSE 8080
ENTRYPOINT ["/server"]
