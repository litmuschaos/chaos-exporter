 
# Makefile for building Chaos Exporter
# Reference Guide - https://www.gnu.org/software/make/manual/make.html

IS_DOCKER_INSTALLED = $(shell which docker >> /dev/null 2>&1; echo $$?)

# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v '/vendor/')

# docker info
DOCKER_REPO ?= litmuschaos
DOCKER_IMAGE ?= chaos-exporter
DOCKER_TAG ?= ci

.PHONY: all
all: format lint deps build test security-checks push 

.PHONY: help
help:
	@echo ""
	@echo "Usage:-"
	@echo "\tmake all   -- [default] builds the chaos exporter container"
	@echo ""

.PHONY: format
format:
	@echo "------------------"
	@echo "--> Running go fmt"
	@echo "------------------"
	@go fmt $(PACKAGES)

.PHONY: lint
lint:
	@echo "------------------"
	@echo "--> Running golint"
	@echo "------------------"
	@golint $(PACKAGES)
	@echo "------------------"
	@echo "--> Running go vet"
	@echo "------------------"
	@go vet $(PACKAGES)

.PHONY: deps
deps: _build_check_docker godeps bdddeps unused-package-check

_build_check_docker:
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

.PHONY: build  
build: go-build docker-build

go-build:
	@echo "------------------"
	@echo "--> Build Chaos Exporter"
	@echo "------------------"
	@bash build/go-multiarch-build.sh ./cmd/exporter 

docker-build: 
	@echo "------------------"
	@echo "--> Build chaos-exporter image" 
	@echo "------------------"
	# Dockerfile available in the repo root
	@sudo docker buildx build --file Dockerfile --progress plane --platform linux/arm64,linux/amd64 --no-cache --tag $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: test
test:
	@echo "------------------"
	@echo "--> Run Go Test"
	@echo "------------------"
	@go test ./... -v -count=1

.PHONY: security-checks
security-checks: trivy-security-check

trivy-security-check:
	@echo "------------------"
	@echo "--> Trivy Security Check"
	@echo "------------------"
	./trivy --exit-code 0 --severity HIGH --no-progress litmuschaos/chaos-exporter:ci
	./trivy --exit-code 1 --severity CRITICAL --no-progress litmuschaos/chaos-exporter:ci

.PHONY: push
push: docker-push

docker-push:
	@echo "------------------"
	@echo "--> Push chaos-exporter image" 
	@echo "------------------"
	REPONAME="litmuschaos" IMGNAME="chaos-exporter" IMGTAG="ci" ./buildscripts/push

unused-package-check:
	@echo "------------------"
	@echo "--> Check unused packages for the chaos-operator"
	@echo "------------------"
	@tidy=$$(go mod tidy); \
	if [ -n "$${tidy}" ]; then \
		echo "go mod tidy checking failed!"; echo "$${tidy}"; echo; \
	fi

.PHONY: build-amd64
build-amd64:
	@echo "-------------------------"
	@echo "--> Build chaos-exporter image" 
	@echo "-------------------------"
	@sudo docker build --file Dockerfile --tag $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) . --build-arg TARGETARCH=amd64

.PHONY: push-amd64
push-amd64:

	@echo "------------------------------"
	@echo "--> Pushing image" 
	@echo "------------------------------"
	@sudo docker push $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG)	
