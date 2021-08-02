# Image URL to use all building/pushing image targets
IMG ?= openkruise/kruise-state-metrics:test
# Platforms to build the image for
PLATFORMS ?= linux/amd64,linux/arm64,linux/arm

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Build manager binary
manager: fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: fmt vet
	go run main.go --kubeconfig ~/.kube/config

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Build the docker image
docker-build: test
	docker build --pull --no-cache . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# Build and push the multiarchitecture docker images and manifest.
docker-multiarch:
	docker buildx build --pull --no-cache --platform=$(PLATFORMS) --push . -t $(IMG)