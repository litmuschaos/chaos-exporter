 
# Makefile for building Chaos Exporter
# Reference Guide - https://www.gnu.org/software/make/manual/make.html

IS_DOCKER_INSTALLED = $(shell which docker >> /dev/null 2>&1; echo $$?)

# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v '/vendor/')

# docker info
DOCKER_REPO ?= litmuschaos
DOCKER_REPO ?= litmuschaos
DOCKER_IMAGE ?= chaos-exporter
DOCKER_TAG ?= ci

.PHONY: unused-package-check
unused-package-check:
	@echo "------------------"
	@echo "--> Check unused packages for the chaos-operator"
	@echo "------------------"
	@tidy=$$(go mod tidy); \
	if [ -n "$${tidy}" ]; then \
		echo "go mod tidy checking failed!"; echo "$${tidy}"; echo; \
	fi

.PHONY: deps
deps: build_check_docker godeps bdddeps

build_check_docker:
	@if [ $(IS_DOCKER_INSTALLED) -eq 1 ]; \
		then echo "" \
		&& echo "ERROR:\tdocker is not installed. Please install it before build." \
		&& echo "" \
		&& exit 1; \
		fi;

godeps:
	@echo ""
	@echo "INFO:\tverifying dependencies for chaos exporter build ..."
	@go get -u -v golang.org/x/lint/golint
	@go get -u -v golang.org/x/tools/cmd/goimports
	

.PHONY: bdddeps
bdddeps:
	@echo "------------------"
	@echo "bdd test dependencies"
	@echo "INFO:\tverifying dependencies for bdddeps ..."
	@echo "------------------"
	@go get -u github.com/onsi/ginkgo
	@go get -u github.com/onsi/gomega 
	kubectl create -f https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/chaos_crds.yaml
	kubectl create ns litmus

.PHONY: test
test:
	@echo "------------------"
	@echo "--> Run Go Test"
	@echo "------------------"
	@go test ./... -v -count=1

.PHONY: build-chaos-exporter build-chaos-exporter-amd64 push-chaos-exporter

publish-chaos-exporter: build-chaos-exporter push-chaos-exporter

build-chaos-exporter:
	@docker buildx build --file Dockerfile --progress plane --platform linux/arm64,linux/amd64 --tag $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) .

build-chaos-exporter-amd64:
	@docker build -f Dockerfile -t $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) .  --build-arg TARGETPLATFORM="linux/amd64"

push-chaos-exporter:
	@docker buildx build --file Dockerfile --progress plane --push --platform linux/arm64,linux/amd64 --tag $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) .
	@docker buildx build --file Dockerfile --progress plane --push --platform linux/arm64,linux/amd64 --tag $(DOCKER_REPO)/$(DOCKER_IMAGE):latest .
