# Image URL to use all building/pushing image targets
IMG ?= openkruise/kruise-state-metrics:test
# Platforms to build the image for
PLATFORMS ?= linux/amd64,linux/arm64

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: build

# Build kruise-state-metrics binary
build: fmt vet
	go build -o bin/kruise-state-metrics main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: fmt vet
	go run main.go --kubeconfig ~/.kube/config

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

lint: golangci-lint ## Run golangci-lint against code.
	$(GOLANGCI_LINT) run

# Build the docker image
docker-build:
	docker build --pull --no-cache . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# Build and push the multiarchitecture docker images and manifest.
docker-multiarch:
	docker buildx build -f ./Dockerfile_multiarch --pull --no-cache --platform=$(PLATFORMS) --push . -t $(IMG)

deploy: kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd deploy && $(KUSTOMIZE) edit set image kruise-state-metrics=${IMG}
	$(KUSTOMIZE) build deploy | kubectl apply -f -
	echo "resources:\n- deploy.yaml" > deploy/kustomization.yaml

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build deploy | kubectl delete -f -

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.2)

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
golangci-lint: ## Download golangci-lint locally if necessary.
	$(call go-get-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
