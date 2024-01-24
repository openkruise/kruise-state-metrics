FROM golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY internal/ internal/
COPY pkg/ pkg/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -mod=vendor -a -o kruise-state-metrics main.go

# Use Ubuntu 20.04 LTS as base image to package the binary
FROM ubuntu:focal
WORKDIR /
COPY --from=builder /workspace/kruise-state-metrics .
ENTRYPOINT ["/kruise-state-metrics"]
