# Copyright The AIGW Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SHELL = /bin/bash

PROJECT_NAME    = github.com/aigw-project/metadata-center
BINARY_NAME     = metadata-center
DOCKER_MIRROR   = m.daocloud.io/
BUILD_IMAGE     ?= $(DOCKER_MIRROR)docker.io/library/golang:1.23-alpine
DOCKER_IMAGE    ?= metadata-center

# use for version update
TIMESTAMP := $(shell date "+%Y%m%d%H%M%S")
GIT_COMMIT := $(shell git rev-parse --short HEAD)
VERSION_FILE := VERSION

GO_MODULES = ./cmd/... ./pkg/...

MOUNT_GOMOD_CACHE = -v $(shell go env GOPATH):/go
ifeq ($(IN_CI), true)
	# Mount go mod cache in the CI environment will cause 'Permission denied' error
	# when accessing files on host in later phase because the mounted directory will
	# have files which is created by the root user in Docker.
	# Run as low privilege user in the Docker doesn't
	# work because we also need root to create /.cache in the Docker.
	MOUNT_GOMOD_CACHE =
endif

.PHONY: build-local
build-local:
	go build -v -o $(BINARY_NAME) $(PROJECT_NAME)/cmd

.PHONY: build
build:
	@echo "Building using Docker image: $(BUILD_IMAGE)"
	@docker run --rm $(MOUNT_GOMOD_CACHE) -v $(PWD):/go/src/$(PROJECT_NAME) -w /go/src/$(PROJECT_NAME) \
		-e GOPROXY \
		$(BUILD_IMAGE) \
		make build-local

.PHONY: run-local
run-local: build-local
	POD_IP=127.0.0.1 ./$(BINARY_NAME) run --config configs/config.toml

.PHONY: unit-test-local
unit-test-local:
	go test -v $(GO_MODULES) -covermode=atomic -coverprofile=coverage.out -coverpkg=$(PROJECT_NAME)/...

.PHONY: unit-test
unit-test:
	@echo "Running unit tests using Docker image: $(BUILD_IMAGE)"
	@docker run --rm $(MOUNT_GOMOD_CACHE) -v $(PWD):/go/src/$(PROJECT_NAME) -w /go/src/$(PROJECT_NAME) \
		-e GOPROXY \
		$(BUILD_IMAGE) \
		make unit-test-local

.PHONY: fmt-go
fmt-go:
	@echo "Formatting Go code..."
	@go fmt $(GO_MODULES)

GOLANGCI_LINT_VERSION = 1.62.2
.PHONY: lint-go
lint-go:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION); \
	golangci-lint run --timeout 10m $(GO_MODULES)

LICENSE_CHECKER_VERSION = 0.6.0
.PHONY: install-license-checker
install-license-checker:
	go install github.com/apache/skywalking-eyes/cmd/license-eye@v$(LICENSE_CHECKER_VERSION)

.PHONY: lint-license
lint-license: install-license-checker
	license-eye header check --config .licenserc.yaml

.PHONY: fix-license
fix-license: install-license-checker
	license-eye header fix --config .licenserc.yaml

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out

.PHONY: build-image
build-image:
	@echo "Building Docker image..."
	@docker build --build-arg BUILD_IMAGE=$(BUILD_IMAGE) -t $(DOCKER_IMAGE):latest .

.PHONY: run-docker
run-docker: docker-build
	docker run --rm -p 8080:8080 -p 8081:8081 -e POD_IP=127.0.0.1 $(DOCKER_IMAGE):latest
