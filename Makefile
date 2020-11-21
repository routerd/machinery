# routerd
# Copyright (C) 2020  The routerd Authors
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

SHELL=/bin/bash
.SHELLFLAGS=-euo pipefail -c

UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

TMP_BASE := .tmp
TMP := $(TMP_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
TMP_BIN = $(TMP)/bin
TMP_VERSIONS := $(TMP)/versions

BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
SHORT_SHA=$(shell git rev-parse --short HEAD)
VERSION?=${BRANCH}-${SHORT_SHA}

PROTOC_VERSION:=v3.13.0
PROTOCGENGO_VERSION:=v1.25.0

export CGO_ENABLED:=0
export GO111MODULE:=on
export GOBIN := $(abspath $(TMP_BIN))
export PATH := $(GOBIN):$(PATH)

# -------------------
# Testing and Linting
# -------------------

test:
	CGO_ENABLED=1 go test -race -v ./...
.PHONY: test

fmt:
	go fmt ./...
.PHONY: fmt

vet:
	go vet ./...
.PHONY: vet

tidy:
	go mod tidy
.PHONY: tidy

verify-boilerplate:
	@go run hack/boilerplate/boilerplate.go \
		-boilerplate-dir hack/boilerplate/ \
		-verbose
.PHONY: verify-boilerplate

pre-commit-install:
	@echo "installing pre-commit hooks using https://pre-commit.com/"
	@pre-commit install
.PHONY: pre-commit-install

# ------------
# Dependencies
# ------------


PROTOC := $(TMP_VERSIONS)/protoc/$(PROTOC_VERSION)
$(PROTOC):
	$(eval PROTOC_TMP := $(shell mktemp -d))
	@mkdir -p $(TMP_BIN)
	@ cd $(PROTOC_TMP); \
		curl -L --fail \
	 	https://github.com/protocolbuffers/protobuf/releases/download/$(PROTOC_VERSION)/protoc-$(shell echo -n $(PROTOC_VERSION) | tail -c +2)-linux-$(UNAME_ARCH).zip \
		-o protoc.zip && \
		unzip -qq protoc.zip
	@mv $(PROTOC_TMP)/bin/protoc $(TMP_BIN)
	@mv $(PROTOC_TMP)/include $(TMP)
	@rm -rf $(PROTOC_TMP)
	@rm -rf $(dir $(PROTOC))
	@mkdir -p $(dir $(PROTOC))
	@touch $(PROTOC)

PROTOCGENGO := $(TMP_VERSIONS)/protoc-gen-go/$(PROTOCGENGO_VERSION)
$(PROTOCGENGO):
	$(eval PROTOCGENGO_TMP := $(shell mktemp -d))
	@cd $(PROTOCGENGO_TMP); go get google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOCGENGO_VERSION)
	@rm -rf $(PROTOCGENGO_TMP)
	@rm -rf $(dir $(PROTOCGENGO))
	@mkdir -p $(dir $(PROTOCGENGO))
	@touch $(PROTOCGENGO)

# ----------
# Generators
# ----------

%.proto: FORCE
	$(eval PROTO_DIR := $(shell dirname $@))
	@ echo generating $@
	@protoc \
		--go_out=$(PROTO_DIR) --go_opt=paths=source_relative \
		-I$(TMP)/include \
		-Iapi/v1 \
		-I$(PROTO_DIR) $@

.PHONY: proto
proto: api/**/*.proto storage/testdata/**/*.proto

FORCE:
