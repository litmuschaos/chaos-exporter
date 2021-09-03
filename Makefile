 
# Makefile for building Chaos Exporter
# Reference Guide - https://www.gnu.org/software/make/manual/make.html

IS_DOCKER_INSTALLED = $(shell which docker >> /dev/null 2>&1; echo $$?)

# docker info
DOCKER_REGISTRY ?= docker.io
DOCKER_REPO ?= litmuschaos
DOCKER_IMAGE ?= chaos-exporter
DOCKER_TAG ?= dev

.PHONY: help
help:
	@echo ""
	@echo "Usage:-"
	@echo "\tmake deps          -- sets up dependencies for image build"
	@echo "\tmake build         -- builds the chaos-exporter binary & docker multi-arch image"
	@echo "\tmake push          -- pushes the chaos-exporter multi-arch image"
	@echo "\tmake build-amd64   -- builds the chaos-exporter binary & docker amd64 image"
	@echo "\tmake push-amd64    -- pushes the chaos-exporter amd64 image"
	@echo ""

.PHONY: all
all: deps gotasks build test push

.PHONY: gotasks
gotasks: unused-package-check

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

.PHONY: build_check_docker
build_check_docker:
	@if [ $(IS_DOCKER_INSTALLED) -eq 1 ]; \
		then echo "" \
		&& echo "ERROR:\tdocker is not installed. Please install it before build." \
		&& echo "" \
		&& exit 1; \
		fi;

.PHONY: godeps
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

.PHONY: build
build:
    
	@echo "-------------------------"
	@echo "--> Build go-runner image" 
	@echo "-------------------------"
	@docker buildx build --file Dockerfile --progress plane  --no-cache --platform linux/arm64,linux/amd64 --tag $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: push
push:

	@echo "------------------------------"
	@echo "--> Pushing image"
	@echo "------------------------------"
	@docker buildx build --file Dockerfile --progress plane --no-cache --push --platform linux/arm64,linux/amd64 --tag $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: build-amd64
build-amd64:

	@echo "------------------------------"
	@echo "--> Build go-runner image" 
	@echo "-------------------------"
	@docker build -f Dockerfile  --no-cache -t $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) .  --build-arg TARGETPLATFORM="linux/amd64"

.PHONY: push-amd64
push-amd64:

	@echo "------------------------------"
	@echo "--> Pushing image" 
	@echo "------------------------------"
	@sudo docker push $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG)
