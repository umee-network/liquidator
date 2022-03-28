#!/usr/bin/make -f

BRANCH         := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT         := $(shell git log -1 --format='%H')
BUILD_DIR      ?= $(CURDIR)/build

###############################################################################
##                                  Version                                  ##
###############################################################################

ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

###############################################################################
##                                   Build                                   ##
###############################################################################

ldflags = -X github.com/umee-network/liquidator/cmd.Version=$(VERSION) \
		  -X github.com/umee-network/liquidator/cmd.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(ldflags)'

install: go.sum
	@echo "--> Installing..."
	go install -mod=readonly $(BUILD_FLAGS) ./...

build: go.sum
	@echo "--> Building..."
	go build -mod=readonly -o $(BUILD_DIR)/ $(BUILD_FLAGS) ./...

build-linux: go.sum
	GOOS=linux GOARCH=amd64 $(MAKE) build

go.sum:
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

clean:
	@echo "--> Cleaning..."
	@rm -rf $(BUILD_DIR)

.PHONY: install build build-linux clean go.sum

###############################################################################
##                              Tests & Linting                              ##
###############################################################################

test:
	@echo "--> Running tests"
	@go test ./...

.PHONY: test

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m

.PHONY: lint
