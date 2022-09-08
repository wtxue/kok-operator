VERSION ?= v0.2.0
# Image URL to use all building/pushing image targets
IMG_REG ?= docker.io/wtxue
IMG_CTL := $(IMG_REG)/kok-operator
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# This repo's root import path (under GOPATH).
ROOT := github.com/wtxue/kok-operator

GO_VERSION := 1.19.0
ARCH     ?= $(shell go env GOARCH)
BUILD_DATE = $(shell date +'%Y-%m-%dT%H:%M:%SZ')
COMMIT    = $(shell git rev-parse --short HEAD)
GOENV    := CGO_ENABLED=0 GOOS=$(shell uname -s | tr A-Z a-z) GOARCH=$(ARCH) GOPROXY=https://goproxy.io,direct
GO       := $(GOENV) go build -tags=jsoniter

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: build

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# generate crd spec and deepcopy
crd: generate manifests

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
#	cd config/manager && kustomize edit set image controller=${IMG}
#	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build:
	docker run --rm -v "$$PWD":/go/src/${ROOT} -v ${GOPATH}/pkg/mod:/go/pkg/mod -w /go/src/${ROOT} golang:${GO_VERSION} make build-local

build: build-local

build-local:
	$(GO) -v -o bin/kok-operator -ldflags "-s -w -X $(ROOT)/pkg/version.Release=$(VERSION) -X  $(ROOT)/pkg/version.Commit=$(COMMIT)   \
	-X  $(ROOT)/pkg/version.BuildDate=$(BUILD_DATE)" cmd/controller/main.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64           \
	go build -tags=jsoniter -v -o bin/kok-operator -ldflags "-s -w -X $(ROOT)/pkg/version.Release=$(VERSION) -X  $(ROOT)/pkg/version.Commit=$(COMMIT)   \
    -X  $(ROOT)/pkg/version.BuildDate=$(BUILD_DATE)" cmd/controller/main.go


# Push the docker image
push:
	docker build -t ${IMG_CTL}:${VERSION} -f ./docker/kok-operator/Dockerfile .
	docker push ${IMG_CTL}:${VERSION}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.9.2 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
