################################################################################
##                             VERSION PARAMS                                 ##
################################################################################

## Docker Build Versions
DOCKER_BUILD_IMAGE = golang:1.16.2
DOCKER_BASE_IMAGE = alpine:3.12.4

## Tool Versions
TERRAFORM_VERSION=0.13.5

################################################################################

GO ?= $(shell command -v go 2> /dev/null)
MATTERMOST_APPS_CLOUD_DEPLOYER_IMAGE ?= mattermost/mattermost-apps-cloud-deployer:test
MACHINE = $(shell uname -m)
GOFLAGS ?= $(GOFLAGS:)
BUILD_TIME := $(shell date -u +%Y%m%d.%H%M%S)
BUILD_HASH := $(shell git rev-parse HEAD)

################################################################################

LOGRUS_URL := github.com/sirupsen/logrus

LOGRUS_VERSION := $(shell find go.mod -type f -exec cat {} + | grep ${LOGRUS_URL} | awk '{print $$NF}')

LOGRUS_PATH := $(GOPATH)/pkg/mod/${LOGRUS_URL}\@${LOGRUS_VERSION}

export GO111MODULE=on

all: check-style dist

## Runs govet and gofmt against all packages.
.PHONY: check-style
check-style: govet lint
	@echo Checking for style guide compliance

## Runs lint against all packages.
.PHONY: lint
lint:
	@echo Running lint
	env GO111MODULE=off $(GO) get -u golang.org/x/lint/golint
	golint -set_exit_status ./...
	@echo lint success

## Runs govet against all packages.
.PHONY: vet
govet:
	@echo Running govet
	$(GO) vet ./...
	@echo Govet success

## Builds and thats all :)
.PHONY: dist
dist:	build

.PHONY: build
build: ## Build the mattermost-apps-cloud-deployer
	@echo Building Mattermos-Apps-Cloud-Deployer
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -gcflags all=-trimpath=$(PWD) -asmflags all=-trimpath=$(PWD) -a -installsuffix cgo -o build/_output/bin/main  ./

build-image:  ## Build the docker image for mattermost-apps-cloud-deployer
	@echo Building Mattermos-Apps-Cloud-Deployer Docker Image
	docker build \
	--build-arg DOCKER_BUILD_IMAGE=$(DOCKER_BUILD_IMAGE) \
	--build-arg DOCKER_BASE_IMAGE=$(DOCKER_BASE_IMAGE) \
	. -f build/Dockerfile -t $(MATTERMOST_APPS_CLOUD_DEPLOYER_IMAGE) \
	--no-cache

.PHONY: install
install: build
	go install ./...

get-terraform: ## Download terraform only if it's not available. Used in the docker build
	@if [ ! -f build/terraform ]; then \
		curl -Lo build/terraform.zip https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && cd build && unzip terraform.zip &&\
		chmod +x terraform && rm terraform.zip;\
	fi