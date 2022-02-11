# Image URL to use all building/pushing image targets
#PADDLEJOB_IMG ?= registry.baidubce.com/paddle-operator/paddlejob
#SAMPLESET_IMG ?= registry.baidubce.com/paddle-operator/sampleset
#RUNTIME_IMG ?= registry.baidubce.com/paddle-operator/runtime
#SERVING_IMG ?= registry.baidubce.com/paddle-operator/serving

PADDLEJOB_IMG ?= xiaolao/paddlejob
SAMPLESET_IMG ?= xiaolao/sampleset
RUNTIME_IMG ?= xiaolao/runtime
SERVING_IMG ?= xiaolao/serving

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:maxDescLen=0,generateEmbeddedObjectMeta=true,trivialVersions=true,preserveUnknownFields=false"

# Set version and get git tag
VERSION=v0.4
GIT_SHA=$(shell git rev-parse --short HEAD || echo "HEAD")
GIT_VERSION=${VERSION}-${GIT_SHA}

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Run tests
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: generate fmt vet manifests
	mkdir -p ${ENVTEST_ASSETS_DIR}
#	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.11.0/hack/setup-envtest.sh
#	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

gen-deploy: manifests kustomize crd-v1beta1
	cat COPYRIGHT.YAML > deploy/v1/crd.yaml
	$(KUSTOMIZE) build config/crd >> deploy/v1/crd.yaml
	cat COPYRIGHT.YAML > deploy/v1/operator.yaml
	$(KUSTOMIZE) build config/operator >> deploy/v1/operator.yaml
	cat COPYRIGHT.YAML > deploy/extensions/controllers.yaml

TEMPLATES_DIR ?= charts/paddle-operator/templates
helm: manifests kustomize
	$(KUSTOMIZE) build config/crd > $(TEMPLATES_DIR)/crd.yaml
	$(KUSTOMIZE) build config/default > $(TEMPLATES_DIR)/controller.yaml
	# The last sed command is not clean, need more advanced scripting to be robust
	sed -i  -e "s/image:.*/image: {{ .Values.image }}/g" \
			-e "s/--namespace=.*/--namespace={{ .Values.jobnamespace }}/g" \
			-e "s/namespace:.*/namespace: {{ .Values.controllernamespace }}/g" \
			-e "s/name: paddle-system/name: {{ .Values.controllernamespace }}/g" \
			$(TEMPLATES_DIR)/controller.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Build all controller manager docker images
docker-build-all: docker-build-paddlejob docker-build-sampleset docker-build-runtime docker-build-serving

# Build paddlejob controller image
docker-build-paddlejob: test
	docker build . -f docker/Dockerfile.paddlejob -t ${PADDLEJOB_IMG}:${GIT_VERSION}

# Build sampleset controller image
docker-build-sampleset: test
	docker build . --build-arg RUNTIME_IMG=${RUNTIME_IMG} --build-arg GIT_VERSION=${GIT_VERSION} \
		-f docker/Dockerfile.sampleset -t ${SAMPLESET_IMG}:${GIT_VERSION}

docker-build-runtime: test
	docker build . -f docker/Dockerfile.runtime -t ${RUNTIME_IMG}:${GIT_VERSION}

docker-build-serving: test
	docker build . -f docker/Dockerfile.serving -t ${SERVING_IMG}:${GIT_VERSION}

# Push all docker images
docker-push-all: docker-push-paddlejob docker-push-sampleset docker-push-runtime docker-push-serving

# Push the docker image
docker-push-paddlejob:
	docker push ${PADDLEJOB_IMG}:${GIT_VERSION}

# Push the sampleset docker image
docker-push-sampleset:
	docker push ${SAMPLESET_IMG}:${GIT_VERSION}

docker-push-runtime:
	docker push ${RUNTIME_IMG}:${GIT_VERSION}

docker-push-serving:
	docker push ${SERVING_IMG}:${GIT_VERSION}

update-api-doc:
	bash docs/api-doc-gen/gen_api_doc.sh

# Download controller-gen locally if necessary
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen:
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.1)

# Download kustomize locally if necessary
KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize:
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
