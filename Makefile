# Makefile for building Chaos Exporter
# Reference Guide - https://www.gnu.org/software/make/manual/make.html

IS_DOCKER_INSTALLED = $(shell which docker >> /dev/null 2>&1; echo $$?)


# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v '/vendor/')

.PHONY: all
all: format lint build test dockerops 

.PHONY: help
help:
	@echo ""
	@echo "Usage:-"
	@echo "\tmake all   -- [default] builds the chaos exporter container"
	@echo ""

.PHONY: godeps
godeps:
	@echo ""
	@echo "INFO:\tverifying dependencies for chaos exporter build ..."
	@go get -u -v golang.org/x/lint/golint
	@go get -u -v golang.org/x/tools/cmd/goimports
	@go get -u -v github.com/golang/dep/cmd/dep

_build_check_docker:
	@if [ $(IS_DOCKER_INSTALLED) -eq 1 ]; \
		then echo "" \
		&& echo "ERROR:\tdocker is not installed. Please install it before build." \
		&& echo "" \
		&& exit 1; \
		fi;

.PHONY: deps
deps: _build_check_docker godeps

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

.PHONY: build  
build:
	@echo "------------------"
	@echo "--> Build Chaos Exporter"
	@echo "------------------"
	@go build ./cmd/exporter 

.PHONY: test
test:
	@echo "------------------"
	@echo "--> Run Go Test"
	@echo "------------------"
	@go test ./... -v 

.PHONY: dockerops
dockerops: 
	@echo "------------------"
	@echo "--> Build chaos-exporter image" 
	@echo "------------------"
	# Dockerfile available in the repo root
	sudo docker build . -f Dockerfile -t litmuschaos/chaos-exporter:ci  
	REPONAME="litmuschaos" IMGNAME="chaos-exporter" IMGTAG="ci" ./buildscripts/push

.PHONY: bdddeps
bdddeps:
	@echo "bdd test dependencies"
	@echo "INFO:\tverifying dependencies for bdddeps ..."
	@go get -u -v github.com/litmuschaos/chaos-exporter/pkg/util
	@go get -u -v github.com/litmuschaos/chaos-exporter/pkg/clientset/v1alpha1
	@go get -u -v github.com/litmuschaos/chaos-operator/pkg/apis
	@go get -u -v github.com/onsi/ginkgo
	@go get -u -v github.com/onsi/gomega 
